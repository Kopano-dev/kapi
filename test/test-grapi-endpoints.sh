#!/bin/sh

set -e

if [ -n "${REFRESH_TOKEN_VALUE}" ]; then
	export "$(${PYTHON:-python3} get-access-token.py --format env | \grep REFRESH_TOKEN_VALUE)"
fi

BASE_URL="${BASE_URL:-https://$(hostname):8428}/api/gc/v1"

. ./keep-access-token.sh >/dev/null

call_endpoint() {
	./curl.sh "${BASE_URL}${1}" 2>/dev/null | jq -r "$2"
}

assert_true() {
	if [ "${1}" = "${2}" ]; then
		printf "\e[92m[PASSED]\e[0m %s == %s\n" "${1}" "${2}"
	else
		printf "\e[91m[FAILED]\e[0m %s != %s (%s)\n" "${1}" "${2}" "${3}"
	fi
}

# test me
assert_true "$(call_endpoint "/me" ".displayName")" "Timmothy Schöwalter" "displayName is incorrect!"
assert_true "$(call_endpoint "/me" ".jobTitle")" "Arts development officer" "jobTitle is incorrect!"
assert_true "$(call_endpoint "/me/calendar" ".name")" "Calendar" ".name is incorrect!"

# test me/contacts
assert_true "$(call_endpoint "/me/contacts" ".value | length")" "0" "number of contactFolders expected to be 0, but it's not!"

# test me/contactFolders
assert_true "$(call_endpoint "/me/contactFolders" ".value | length")" "2" "number of contactFolders expected to be 2, but it's not!"
test_contact_folder=$(call_endpoint "/me/contactFolders" ".value[1].id")
assert_true "$(call_endpoint "/me/contactFolders/${test_contact_folder}" ".displayName")" "Suggested Contacts" "contactFolders name is incorrect!"

# test me/calendars
assert_true "$(call_endpoint "/me/calendars" ".value | length")" "1" "number of calendars expected to be 1, but it's not!"
test_calendar=$(call_endpoint "/me/calendars" ".value[0]")
test_calendar_id="$(jq -r .id <<< "${test_calendar}")"
assert_true "$(jq -r .name <<< "${test_calendar}")" "Calendar" "calendar name is incorrect!"

# test me/events
assert_true "$(call_endpoint "/me/events" ".value | length")" "0" "number of events expected to be 0, but it's not!"
assert_true "$(call_endpoint "/me/calendars/${test_calendar_id}/events" ".value | length")" "0" "number of events in the calendar expected to be 0, but it's not!"

# test me/photo
test_photo_id="$(call_endpoint "/me/photo" ".id")"
assert_true "${test_photo_id}" "612X408" "Photo ID is not equal!"
assert_true "$(call_endpoint "/me/photo/${test_photo_id}" ".id")" "612X408" "Photo ID is not equal!"
assert_true "$(call_endpoint "/me/photos" ".value | length")" "1" "number of photos expected to be '1', but it's not!"

# test me/mailFolders
assert_true "$(call_endpoint "/me/mailFolders" ".value | length")" "6" "number of mailFolders expected to be '6', but it's not!"
test_folder="$(call_endpoint "/me/mailFolders" ".value[0]")"
test_folder_id="$(jq -r ".id" <<< "${test_folder}")"
assert_true "$(jq -r .displayName <<< "${test_folder}")" "Inbox" "First folder expected to be 'Inbox', but it's not!"
assert_true "$(call_endpoint "/me/mailFolders/${test_folder_id}" ".displayName")" "Inbox" "Folder displayName expected to be 'Inbox', but it's not!"

assert_true "$(call_endpoint "/me/messages" ".value | length")" "0" "number of messages expected to be '0', but it's not!"
assert_true "$(call_endpoint "/me/mailFolders/${test_folder_id}/messages" ".value | length")" "0" "number of messages expected to be '0', but it's not!"

# test groups
assert_true "$(call_endpoint "/groups" ".value | length")" "7" "number of groups expected to be 7, but it's not!"

# test a user
assert_true "$(call_endpoint "/users" ".value | length")" "10" "number of users expected to be 10, but it's not!"
test_user=$(call_endpoint "/users" ".value[1].id")
assert_true "$(call_endpoint "/users/${test_user}" ".displayName")" "Alford Predovic" "displayName is incorrect!"
assert_true "$(call_endpoint "/users/${test_user}" ".jobTitle")" "Training and development officer" "jobTitle is incorrect!"
#assert_true "$(call_endpoint "/users/${test_user}/calendar" ".name")" "Calendar" ".name is incorrect!"

# FIXME (mort), this part needs to be updated when grapi's bugs have been fixed!
# test me/photo
#test_photo_id="$(call_endpoint "/users/${test_user}/photo" ".id")"
#assert_true "${test_photo_id}" "612X408" "Photo ID is not equal!"
#assert_true "$(call_endpoint "/users/${test_user}/photo/${test_photo_id}" ".id")" "612X408" "Photo ID is not equal!"
assert_true "$(call_endpoint "/users/${test_user}/photos" ".value|length")" "0" "number of photos expected to be 0, but it's not!"
