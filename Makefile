#/*******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************/

.PHONY: block-storage-attacher
block-storage-attacher:
	cd block-storage-attacher; \
	make vet; \
	make fmt; \
	make test; \
	make coverage

