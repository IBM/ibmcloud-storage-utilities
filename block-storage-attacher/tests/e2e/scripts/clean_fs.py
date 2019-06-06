#!/usr/bin/env python
from __future__ import print_function
import SoftLayer
import argparse
import sys
import os
import logging
import sys, getopt
import json, csv
from datetime import datetime

logger = logging.getLogger()
result = {}
deleted_fs_count = 0
count = 0
number_of_fs_to_be_considered = 0

def is_vm_in_authorized_hosts(vm, authorized_hosts):
    print("entry slcient.is_vm_in_authorized_hosts()")
    # if authorized_hosts is empty, then all hosts are authorized
    if not authorized_hosts or authorized_hosts[0] == '':
        return 1

    for host in authorized_hosts:
        if vm['hostname'] == host:    # VM found
            return 1
    # VM not found
    return 0

# Check if a VM is already allowed to access this share
def is_vm_authorized(vm, allowed_vms_list):
    print("entry slcient.is_vm_authorized()")
    if not allowed_vms_list:
        return 0

    for host in allowed_vms_list:
        if vm['id'] == host['id']:
            return 1

    return 0

def get_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument('username', help='SL Storage API username [prod-lon02-containers.alchemy|prod-dal09-containers.alchemy|stage.containers.alchemy|containers-devtest]')
    parser.add_argument('password', help='SL Storage API password')
    return parser.parse_args()

def remove_storage_access_to_hosts(client, storage_entry, hosts):
    try:
        print("entry EnduranceStorageProvider.remove_storage_access_to_hosts()")
        # Get the list of VMs that already have access to this storage
        allowed_vms_list = client['Network_Storage'].getAllowedVirtualGuests(id = storage_entry['id'])
        # Get the list of all visible VMs
        vm_list = client['Account'].getVirtualGuests()
        #print(vm_list)

        # Remove access to this storage for VMs
        for vm in vm_list:
            if is_vm_in_authorized_hosts(vm,hosts) == 1:
                if is_vm_authorized(vm, allowed_vms_list) == 1:
                    print("Host " + vm['hostname'] + " remove access.")
                    # This VM needs to be Removed access
                    vm_client = client['Virtual_Guest']
                    print("Removing access to VM with hostname " + vm['hostname'])
                    try:
                        result = vm_client.removeAccessToNetworkStorage(storage_entry, id = vm['id'])
                        print("Result = " + str(result))
                    except SoftLayer.SoftLayerAPIError as e:
                        print("Unable to retrieve information faultCode=%s, faultString=%s"
                                     % (e.faultCode, e.faultString))
                        pass
                else:
                    continue

        # Do the same for all physical servers - for now, not accepting list of servers, removing access to all
        allowed_hosts_list = client['Network_Storage'].getAllowedHardware(id = storage_entry['id'])
        # Get the list of all visible hardware servers
        hardware_mask = 'id, primaryIpAddress, hostname'
        hardware_list = client['Account'].getHardware(mask = hardware_mask)
        #print(hardware_list)

        # Remove access to this storage for all physical servers
        for hardware in hardware_list:
            # First check if it already has access - using the same method as VMs as its just check in a list
            if is_vm_authorized(hardware, allowed_hosts_list) == 1:
                #print("Host " + hardware['hostname'] + " has access.")

                # This server needs to be removed access
                hardware_client = client['Hardware_Server']
                print("Removing access to server with hostname " + hardware['hostname'])
                try:
                    result = hardware_client.removeAccessToNetworkStorage(storage_entry, id = hardware['id'])
                    print("Result = " + str(result))
                except SoftLayer.SoftLayerAPIError as e:
                    print("Unable to retrieve information faultCode=%s, faultString=%s"
                                 % (e.faultCode, e.faultString))
                    pass
            else:
                continue


    except SoftLayer.SoftLayerAPIError as e:
        print("Unable to retrieve information faultCode=%s, faultString=%s"
                % (e.faultCode, e.faultString))
        pass

def remove_storage_access_to_hosts_by_subnet(client, storage_entry):
    global deleted_fs_count
    allowed_subnet_list = client['Network_Storage'].getAllowedSubnets(id = storage_entry['id'])
    for subnet in allowed_subnet_list:
        result = client['Network_Subnet'].removeAccessToNetworkStorageList(storage_entry, id = subnet['id'])
        print("Result = " + str(result))
        if str(result) == 'True':
            deleted_fs_count += 1
    print("exit EnduranceStorageProvider.remove_storage_access_to_hosts_by_subnet()")

def delete_file_share(user, pwd, orderid):
    client = SoftLayer.Client(username=user, api_key=pwd)
    #print "Using username = %s , orderid = %s" % (user ,orderid)
    storage_volume_id = int(orderid)
    #print "deleteing order="+str(storage_volume_id)
    objectFilter = {'networkStorage': {'billingItem': {'orderItem': {'order': {'userRecord': {'username': {'operation': user}}}}}}}
    storage_mask = 'id'
    storage_units = client['Account'].getNetworkStorage(filter=objectFilter, mask = storage_mask)
    storage_entry = None
    for storage_unit in storage_units:
        # Find a match for the order_item_id
        entry_orderId = storage_unit['id']
        if entry_orderId == storage_volume_id:
            # Found the storage unit
            storage_entry = storage_unit
            break
    if storage_entry is None:
        print(" Storage is not found with order id ="+orderid)
        return


    #confirm = raw_input(" Delete orderid = %s ? y/n" % storage_volume_id)
    confirm = "y"
    print("input = %s " % confirm)
    if confirm == "y" or confirm == "Y":
        hosts = []
        remove_storage_access_to_hosts(client, storage_entry, hosts)
        remove_storage_access_to_hosts_by_subnet(client, storage_entry)

        billing_mask = 'id'
        units = client['Network_Storage'].getBillingItem(id = storage_volume_id, mask = billing_mask)
        print(units)
        result = client['Billing_Item'].cancelService('False', 'False', "No longer needed", id = units['id'])
        print(result)

def main():
    args = get_arguments()
    user=args.username
    pwd=args.password
    client = SoftLayer.Client(username=user, api_key=pwd)
    print("Using username = %s " % user)
    client = SoftLayer.Client(username=user, api_key=pwd)
    storage_mask = 'id, capacityGb, username, createDate, billingItem[id,orderItemId]'
    units = client['Account'].getNetworkStorage(mask = storage_mask)

    #print units
    #print json.dumps(units, indent=4, sort_keys=True)
    global deleted_fs_count
    global number_of_fs_to_be_considered
    global count
    with open("all_fileshares_of_account.csv", "wb+") as file:
        csv_file = csv.writer(file)
        for item in units:
            if count == 0:
                header = item.keys()
                csv_file.writerow(header)
            count += 1
            csv_file.writerow(item.values())

    print("Will delete the file shares of %s older than 2 weeks" % user)
    deleted_fs = []
    for entry in units:
        created_date = str(entry['createDate'])
        end = created_date.find('T')
        old_date = datetime.strptime(created_date[0:end], "%Y-%m-%d")
        today = datetime.strptime(datetime.now().strftime("%Y-%m-%d"), "%Y-%m-%d")
        delta = today - old_date
        if delta.days >= 14:
            deleted_fs.append(entry)
            number_of_fs_to_be_considered += 1
            print("=================================================================================================")
            print(" Username: " + entry['username'])
            print(" A storage account's capacity, measured in gigabytes: " + str(entry['capacityGb']))
            print(" The date a network storage volume was created: " + str(entry['createDate']))
            print(" SoftLayer Order ID: " + str(entry['id']))
            delete_file_share(user, pwd, str(entry['id']))

    with open("fileshares_considered_for_deletion.csv", "wb+") as file:
        csv_file = csv.writer(file)
        csv_file.writerow(header)
        for item in deleted_fs:
            csv_file.writerow(item.values())

    print("")
    print("==============================================Summary============================================")
    print("Using username = %s " % user)
    print("Total number of file shares: ", count)
    print("Total file shares older than 2 weeks: ", number_of_fs_to_be_considered)
    print("Total number of deleted file shares: ", deleted_fs_count)
    print("")
    print("==============================================Summary============================================")
    print("Fileshares considered for deletion (already deleted fileshares will be ignored):")
    #print json.dumps(deleted_fs, indent=4, sort_keys=True)
    with open('fileshares_considered_for_deletion.csv', 'rb') as f:
        reader = csv.reader(f)
        for row in reader:
            print(row)
    print("=================================================================================================")

if __name__=="__main__":
    print("START")
    main()
