"""
This is a simple Molotov https://molotov.readthedocs.io/ script to bang
the Kopano API server with a Kopano Calendar scenario.
"""

from datetime import datetime, timedelta
import dateutil.relativedelta
import dateutil.parser
import os
import secrets
import time
import tzlocal

import molotov


_API = os.environ.get('KAPID_URL', 'http://127.0.0.1:8039')
_GCAPI = '%s/api/gc/v1' % _API
_HEADERS = {
    'X-Request-With-Molotov': '1'
}
_ACCESS_TOKEN = os.environ.get('TOKEN_VALUE')
_NOW = os.environ.get('NOW')
_LOCAL_TZ = tzlocal.get_localzone().zone

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


def last_day_of_month(any_day):
    next_month = any_day.replace(day=28) + timedelta(days=4)  # this will never fail
    return next_month - timedelta(days=next_month.day)


@molotov.global_setup()
def _setup(args):
    _HEADERS['Authorization'] = "Bearer %s" % _ACCESS_TOKEN
    calendar = molotov.json_request("%s/me/calendar" % _GCAPI, headers=_HEADERS)['content']
    molotov.set_var('calendar', calendar)
    if _NOW:
        now = dateutil.parser.parse(_NOW)
    else:
        now = datetime.now()
    print("\nBase datetime: %s (%s)\n" % (now, _LOCAL_TZ))
    molotov.set_var('now', now)


@molotov.global_teardown()
def display_average():
    average = sum(_R.values()) / len(_R)
    print("\nAverage response time %.5fms" % average)


@molotov.setup()
async def init_worker(worker_num, args):
    return {
        'headers': _HEADERS
    }


@molotov.scenario(weight=100)
async def scenario_get_gc_me_calendar(session):
    calendarID = molotov.get_var('calendar')['id']
    async with session.get("%s/me/calendar" % _GCAPI) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/api/gc/v1/me/calendar')
        assert res.get('id') == calendarID


async def sub_scenario_get_gc_me_calendar_events_by_id(session, eventID):
    async with session.get("%s/me/calendar/events/%s" % (_GCAPI, eventID)) as resp:
        assert resp.status in (200, 404), "HTTP response status: %d" % resp.status
        if resp.status == 200:
            res = await resp.json()
            assert res.get('@odata.context', '').endswith('/me/calendar/events/%s' % eventID)
            assert res.get('id') == eventID

async def sub_scenario_delete_gc_me_calendar_events_by_id(session, eventID):
    async with session.delete("%s/me/calendar/events/%s" % (_GCAPI, eventID)) as resp:
        assert resp.status in (204, 404), "HTTP response status: %d" % resp.status


async def sub_scenario_get_gc_me_calendar_events_all(session, calendarValues):
    for entry in calendarValues:
        eventID = entry.get('id')
        if eventID:
            await sub_scenario_get_gc_me_calendar_events_by_id(session, eventID=eventID)


@molotov.scenario(weight=25)
async def scenario_get_gc_me_calendar_calendarView_3_months(session):
    calendarID = molotov.get_var('calendar')['id']
    now = molotov.get_var('now')
    startDateTime = now + dateutil.relativedelta.relativedelta(day=1, hour=0, minute=0, second=0, microsecond=0,  months=-1)
    endDateTime = last_day_of_month(now.replace(day=28, hour=23, minute=59, second=50, microsecond=999) + timedelta(days=4))
    async with session.get("%s/me/calendars/%s/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,isAllDay,start,end" % (_GCAPI, calendarID, startDateTime.isoformat(), endDateTime.isoformat())) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/me/calendars/%s/calendarView' % calendarID)
        value = res.get('value')
        if value:
            await sub_scenario_get_gc_me_calendar_events_all(session, calendarValues=value)


@molotov.scenario(weight=50)
async def scenario_get_gc_me_calendar_calendarView_3_weeks(session):
    calendarID = molotov.get_var('calendar')['id']
    now = molotov.get_var('now')
    startDateTime = now + dateutil.relativedelta.relativedelta(hour=0, minute=0, second=0, microsecond=0,  weeks=-1, weekday=0)
    endDateTime = now + dateutil.relativedelta.relativedelta(hour=23, minute=59, second=50, microsecond=999, weeks=1, weekday=6)
    async with session.get("%s/me/calendars/%s/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,isAllDay,start,end" % (_GCAPI, calendarID, startDateTime.isoformat(), endDateTime.isoformat())) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/me/calendars/%s/calendarView' % calendarID)
        value = res.get('value')
        if value:
            await sub_scenario_get_gc_me_calendar_events_all(session, calendarValues=value)


@molotov.scenario(weight=100)
async def scenario_get_gc_me_calendar_calendarView_3_days(session):
    calendarID = molotov.get_var('calendar')['id']
    now = molotov.get_var('now')
    startDateTime = now + dateutil.relativedelta.relativedelta(hour=0, minute=0, second=0, microsecond=0,  days=-1)
    endDateTime = now + dateutil.relativedelta.relativedelta(hour=23, minute=59, second=50, microsecond=999, days=1)
    async with session.get("%s/me/calendars/%s/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,isAllDay,start,end" % (_GCAPI, calendarID, startDateTime.isoformat(), endDateTime.isoformat())) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/me/calendars/%s/calendarView' % calendarID)
        value = res.get('value')
        if value:
            await sub_scenario_get_gc_me_calendar_events_all(session, calendarValues=value)


@molotov.scenario(weight=20)
async def scenario_post_me_calendar_events(session):
    now = molotov.get_var('now')
    startOffset = secrets.randbelow(1440) - 720  # Hours offset from now.
    minute = secrets.choice((0, 10, 15, 20, 30, 40, 45, 50))
    duration = secrets.choice((10, 30, 45, 60, 90, 120))
    startDateTime = now + dateutil.relativedelta.relativedelta(hours=startOffset, minute=minute, second=0, microsecond=0)
    endDateTime = startDateTime + dateutil.relativedelta.relativedelta(minutes=duration)

    data = {
        "subject": "Molotov %s-%d" % (startDateTime.isoformat(), duration),
        "start": {
            "dateTime": startDateTime.isoformat(),
            "timeZone": _LOCAL_TZ
        },
        "end": {
            "dateTime": endDateTime.isoformat(),
            "timeZone": _LOCAL_TZ
        }
    }

    async with session.post("%s/me/calendar/events" % (_GCAPI), json=data) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/api/gc/v1/me/calendar/events')


@molotov.scenario(weight=20)
async def scenario_delete_gc_me_calendar_calendarView_3_weeks_molotov_prefix(session):
    calendarID = molotov.get_var('calendar')['id']
    now = molotov.get_var('now')
    startDateTime = now + dateutil.relativedelta.relativedelta(day=1, hour=0, minute=0, second=0, microsecond=0,  months=-1)
    endDateTime = last_day_of_month(now.replace(day=28, hour=23, minute=59, second=50, microsecond=999) + timedelta(days=4))
    async with session.get("%s/me/calendars/%s/calendarView?startDateTime=%s&endDateTime=%s&$select=subject,isAllDay,start,end" % (_GCAPI, calendarID, startDateTime.isoformat(), endDateTime.isoformat())) as resp:
        assert resp.status == 200, "HTTP response status: %d" % resp.status
        res = await resp.json()
        assert res.get('@odata.context', '').endswith('/me/calendars/%s/calendarView' % calendarID)
        value = res.get('value')
        if value:
            for entry in value:
                subject = entry.get('subject')
                if subject.startswith('Molotov '):
                    if secrets.randbelow(100) < 10:
                        await sub_scenario_delete_gc_me_calendar_events_by_id(session, eventID=entry.get('id'))
