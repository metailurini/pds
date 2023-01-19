name="$1"
details=$(echo "$2" | \
  python3 -c "
import sys, json

data = json.load(sys.stdin)
containerStatuses = data.get('containerStatuses')

if containerStatuses != None:
  for containerStatus in containerStatuses:
    state = containerStatus.get('state')

    if state != None:
      waiting = state.get('waiting')

      if waiting != None:
        print(f\"{waiting['reason']} - {waiting['message']}\")
")

if [ -n "$details" ]
then
  echo "$name: $details" >> /tmp/k8s_client_log
  paplay /usr/share/sounds/freedesktop/stereo/message-new-instant.oga
fi
