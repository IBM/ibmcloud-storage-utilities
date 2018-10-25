#!/usr/bin/env python3.6
#
#  MkPVYaml :  Generate PV yaml files for iSCSI external vols on IKS
#
#  Goal:  Autogenerate the "pv.yaml" files needed for worker nodes,
#         as documented here: https://github.com/akgunjal/block-volume-attacher
#
#  Requirements:
#        Ensure the following environment variables are set:
#          SL_API_KEY
#          SL_USERNAME
#          SL_API_KEY
#        Make sure to do "bx login"
#
#  Input file:
#        Assumes an input descriptor file named "yamlgen.yaml"
#        of the following format:
#
#          cluster:  jeffpx1               #  name of IKS cluster
#          type:  endurance                #  performance | endurance
#          offering: storage_as_a_service  #  storage_as_a_service
#                                          # | enterprise | performance
#          # performance:
#          # - iops:  100                  # INTEGER between 100 and 1000
#                                          # in multiples of 100
#          endurance:
#          - tier:  0.25                #   [0.25|2|4|10]
#          size:  [ 30 ]                #   Array of disk capacity sizes (ToDo)
#
#  Output:
#         This will create a set of "pv.yaml" files
#         as input to the block-attach daemonset.
#
#  ToDo:
#         - Only tested "storage_as_a_service" offering
#

import sys
import os
from yaml import load
from px_iks_utils import IKS_clusters, IKS_vols
from pprint import pprint as pp


def usage():
    print("Usage: mkpvyaml clustername")
    sys.exit(-1)


def check_cfg(doc):
    # Sanity:  Must have 'type' and 'size' and ('performance' or 'endurance')

    if not ('type' in doc and 'size' in doc and 'cluster'):
        raise Exception("Must have 'cluster', 'type' and 'size'")

    if doc['type'] == 'performance' and 'performance' not in doc:
        raise Exception("'performance' type but no 'performance' clause")

    if doc['type'] == 'endurance' and 'endurance' not in doc:
        raise Exception("'endurance' type but no 'endurance' clause")

    if 'performance' in doc and 'endurance' in doc:
        raise Exception("Must specify 'performance' OR 'endurance'")


#
# Make sure environment has everything needed
def check_env():
    if not (os.getenv("IAMTOKEN")):
        raise Exception("IAMTOKEN not defined")
    if not (os.getenv("SL_USERNAME")):
        raise Exception("SL_USERNAME not defined")
    if not (os.getenv("SL_API_KEY")):
        raise Exception("SL_API_KEY not defined")


#
# Generate the yaml file, needed to attach a given volume to a given worker
#
def mkpvyaml(pv, OutputFile):

    for i in range(len(pv['vols'])):
        print("""
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: %s-pv%s
  annotations:
    ibm.io/iqn: "%s"
    ibm.io/username: "%s"
    ibm.io/password: "%s"
    ibm.io/targetip: "%s"
    ibm.io/lunid: "%s"
    ibm.io/nodeip: "%s"
spec:
  capacity:
    storage: %sGi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /
  storageClassName: ibmc-block-attacher

""" %
              (pv['id'], i + 1, pv['vols'][i]['iqn'],
               pv['vols'][i]['username'], pv['vols'][i]['password'],
               pv['vols'][i]['targetip'],
               pv['vols'][i]['lunid'], pv['vols'][i]['nodeip'],
               pv['vols'][i]['capacity']), file=OutputFile)

# --------------------------------------------------------------------------------
#
#  MAIN routine
#
# ---------------------------------------------------------------------------------


ParamFile = "/data/yamlgen.yaml"
with open(ParamFile, 'r') as f:
    doc = load(f)
check_cfg(doc)

check_env()
IAMToken = os.getenv("IAMTOKEN")

# print ("Cluster = ", doc['cluster'])
# print ("Region = ", doc['region'])
# print ("IAMToken = ", IAMToken)

#
# Only going to be one cluster
#
c = IKS_clusters(IAMToken=IAMToken, region=doc[
                 'region'], cluster=doc['cluster']).list()[0]
if not c:
    print("No cluster named : %s exists in region : %s" %
          (doc['cluster'], doc['region']))
    sys.exit(-1)

print(c['name'],  c['region'])

for w in c['workers']:
    print("        ",  w['id'], w['privateIP'], w['location'])
    w.update({'vols': []})
    for j in doc['size']:
        vol = {}
        vol.update({'size': j})
        if doc['type'] == 'performance':
            iops = doc['performance'][0]['iops']
            service_offering = "performance"
            tier = ""
        else:
            tier = doc['endurance'][0]['tier']
            service_offering = "enterprise"
            iops = ""
        print("Creating Vol of size", j, "with type: ", doc['type'])
        try:
            iops_param = iops if doc['type'] == 'performance' else None
            #
            # order the volume
            #
            print("Ordering block storage of size: %s for host: %s" %
                  (j, w['id']))
            orderId = IKS_vols().order_vol(
                           storage_type=doc['type'],
                           location=w['location'],
                           size=j,
                           tier_level=tier,
                           iops=iops_param,
                           service_offering=service_offering)
            print("ORDER ID = ", orderId)
            vol.update({'orderId': orderId})
            w['vols'].append(vol)
        except RuntimeError:
            raise Exception("IKS_vols.order_vol failed")


for w in c['workers']:
    for j in w['vols']:
        volId = IKS_vols().wait4_vol(j['orderId'])
        print("Order ID = ", j['orderId'], "has created VolId = ", volId)
        access_info = IKS_vols().authorize_host_vol(volId, w['privateIP'])
        Vol = IKS_vols().get_vol(volId)
        j.update({'volId': volId,
                  'iqn': Vol.acls[0]['iqn'],
                  'username': Vol.acls[0]['username'],
                  'password': Vol.acls[0]['password'],
                  'nodeip': Vol.acls[0]['hostIP'],
                  'targetip': Vol.detail['targetIP'],
                  'lunid': Vol.detail['lunID'],
                  'capacity': Vol.detail['capacityGb']})

OutFile = "/data/pv-" + doc['cluster'] + ".yaml"
OutF = open(OutFile, "w")
# OutF = sys.stdout
for w in c['workers']:
    mkpvyaml(w, OutF)
OutF.close()
print("Output file created as : ", OutFile)
