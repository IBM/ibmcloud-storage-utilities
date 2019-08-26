#!/usr/bin/env python3

import yaml
from yaml import load, Loader

f = open('OPENSOURCE', 'w')

with open("glide.lock", 'r') as stream:
    try:
        data = yaml.load(stream, Loader=Loader)
        for dep in data["imports"]:
            f.write(dep["name"] + "," + dep["version"] + '\n')
    except yaml.YAMLError as exc:
        print(exc)

f.close()
