#!/usr/bin/env python

import yaml

f = open('OPENSOURCE', 'w')

with open("glide.lock", 'r') as stream:
    try:
        data = yaml.load(stream, Loader=yaml.Loader)
        for dep in data["imports"]:
            f.write(dep["name"] + "," + dep["version"] + '\n')
    except yaml.YAMLError as exc:
        print(exc)

f.close()
