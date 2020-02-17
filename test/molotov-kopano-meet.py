"""
This is a simple Molotov https://molotov.readthedocs.io/ script to bang
the Kopano API server with a Kopano Meet scenario.
"""

import os
import time

import molotov


_API = os.environ.get('KAPID_URL', 'http://127.0.0.1:8039')
_GCAPI = '%s/api/gc/v1' % _API
_HEADERS = {
    'X-Request-With-Molotov': '1'
}
_ACCESS_TOKEN = os.environ.get('TOKEN_VALUE')


_S = {}
_R = {}


def _now():
    return time.monotonic()


@molotov.events()
async def record_time(event, **info):
    req = info.get('request')
    if event == 'sending_request':
        _S[req] = _now()
    elif event == 'response_received':
        _R[req] = _now() - _S[req]
        del _S[req]


@molotov.global_setup()
def init_test(args):
    _HEADERS['Authorization'] = "Bearer %s" % _ACCESS_TOKEN
    molotov.json_request("%s/me" % _GCAPI, headers=_HEADERS)['content']


@molotov.global_teardown()
def display_average():
    average = sum(_R.values()) / len(_R)
    print("\nAverage response time %.5fms" % average)


@molotov.setup()
async def init_worker(worker_num, args):
    return {'headers': _HEADERS}


@molotov.scenario(weight=100)
async def scenario_gc_users_top_1(session):
    async with session.get("%s/users?$top=1&$skip=100&$select=id" % _GCAPI) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/api/gc/v1/users')
        value = res.get('value')
        assert type(value) is list

        if len(value) == 0:
            async with session.get("%s/users?$top=100&$skip=0" % _GCAPI) as resp:
                assert resp.status == 200, "HTTP response status: %d" % resp.status
                res = await resp.json()
                assert res.get('@odata.context', '').endswith('/api/gc/v1/users')
                value = res.get('value')
                assert type(value) is list
                assert len(value) > 0


@molotov.scenario(weight=200)
async def scenario_gc_me(session):
    async with session.get("%s/me" % _GCAPI) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/me')
        assert res.get('userPrincipalName', '') != ''
