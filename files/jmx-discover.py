#!/usr/bin/python -u

import sys
import re
import json

try:
  values = json.load(sys.stdin)['value']
  if not isinstance(values, list): sys.exit(1)
except:
  sys.exit(1)

def process(bean):
  desc = {'{#JMX.BEAN}': bean.replace(',','+').replace('"','%22')}
  for (key,val) in [item.split('=') for item in re.sub(r'^[^:]+:', '', bean).split(',')]:
    desc['{#JMX.BEAN.' + key.upper() + '}'] = val
  return desc

result = {'data': [process(bean) for bean in values]}
json.dump(result, sys.stdout)
