"""Run a temporary Go TLS process and verify framing/auth without logging credentials."""
import argparse
import base64
import hashlib
import hmac
import http.client
import json
import os
from pathlib import Path
import socket
import ssl
import subprocess
import time


class PinnedConnection(http.client.HTTPSConnection):
    def connect(self):
        raw = socket.create_connection((self.host, self.port), timeout=5)
        self.sock = self._context.wrap_socket(raw, server_hostname='api.braincloudservers.com')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--executable', required=True)
    parser.add_argument('--config', required=True)
    args = parser.parse_args()
    config_path = Path(args.config).resolve()
    config = json.loads(config_path.read_text(encoding='utf-8-sig'))
    if config.get('enableLabAuth') is not True or config.get('nextendoAuth'):
        raise RuntimeError('This smoke probe requires an explicitly enabled lab-only config; it does not certify Nextendo login.')
    encode = lambda value: base64.urlsafe_b64encode(value).decode().rstrip('=')
    subject = config['allowedSubjects'][0]
    claims = {'iss': 'uch-local-lab', 'aud': '0100FCF002A58000', 'sub': subject,
              'exp': int(time.time()) + 60}
    payload = encode(json.dumps(claims, separators=(',', ':')).encode())
    token = payload + '.' + encode(hmac.new(bytes.fromhex(config['labSigningKeyHex']),
                                            payload.encode(), hashlib.sha256).digest())
    context = ssl.create_default_context(cafile=config['certificatePem'])
    flags = subprocess.CREATE_NO_WINDOW if os.name == 'nt' else 0
    process = subprocess.Popen([str(Path(args.executable).resolve()), '-config', str(config_path),
                                '-addr', '127.0.0.2:8443'], creationflags=flags,
                               stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
    try:
        for _ in range(50):
            try:
                connection = PinnedConnection('127.0.0.2', 8443, context=context)
                connection.request('GET', '/health')
                response = connection.getresponse()
                assert response.status == 200
                assert json.loads(response.read())['title'] == '0100FCF002A58000'
                connection.close()
                break
            except OSError:
                if process.poll() is not None:
                    raise RuntimeError('Temporary Go server failed to start.')
                time.sleep(.1)
        else:
            raise RuntimeError('Temporary Go server did not become ready.')

        def dispatch(message, session=None):
            packet = {'packetId': 1, 'messages': [message]}
            if session:
                packet['sessionId'] = session
            connection = PinnedConnection('127.0.0.2', 8443, context=context)
            connection.request('POST', '/dispatcherv2', json.dumps(packet),
                               {'Content-Type': 'application/json'})
            response = connection.getresponse()
            assert response.status == 200
            result = json.loads(response.read())
            connection.close()
            assert result['packetId'] == 1
            return result['responses'][0]

        auth = {'service': 'authenticationV2', 'operation': 'AUTHENTICATE',
                'data': {'authenticationType': 'Nintendo', 'externalId': subject,
                         'authenticationToken': token}}
        login = dispatch(auth)
        assert login['status'] == 200 and login['data']['identity']['identityData'] == {}
        session = login['data']['sessionId']
        script = {'service': 'script', 'operation': 'RUN',
                  'data': {'scriptName': 'events/getFrozenLobby', 'scriptData': {}}}
        assert dispatch(script, session)['data']['response']['scriptData']['frozenCode'] == ''
        assert dispatch({'service': 'unknown', 'operation': 'unknown'}, session)['reason_code'] == 40333
        auth['data']['authenticationToken'] = 'invalid.token'
        assert dispatch(auth)['reason_code'] == 40307
        assert dispatch({'service': 'playerState', 'operation': 'LOGOUT'}, session)['status'] == 200
        assert dispatch(script, session)['reason_code'] == 40304
        print('Go TLS: certificate/hostname, HTTP framing, authentication, script envelope, rejection and logout passed.')
    finally:
        process.terminate()
        process.wait(timeout=10)


if __name__ == '__main__':
    main()
