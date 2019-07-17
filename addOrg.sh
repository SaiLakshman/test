#create a folder for new org and place the new configtx.yaml and crypto-config.yaml for third organisation

#copy the orderer cryptos to the newly generated cryptos of org3

#for creating the json form of org3 
export FABRIC_CFG_PATH=$PWD && ../../bin/configtxgen -printOrg Org3MSP > ../channel-artifacts/org3.json

#fetching the config block in protobuf format
peer channel fetch config config_block.pb -o orderer.test:7050 -c testcommon --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem 

#convert config_block.pb into json format and extract the config data and keep it in config.json
configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json

#adding Org3MSP to the config.json file and storing it in modified_config.json
jq -s '.[0] * {"channel_group":{"groups":{"Application":{"groups": {"Org3MSP":.[1]}}}}}' config.json ./channels/org3.json > modified_config.json


#finding the difference between two config json and modified config json and storing it in modified_config_pb
#for that first we need to encode the json to .pb and find the diff and store it in .pb
#converting config.json to config.pb
configtxlator proto_encode --input config.json --type common.Config --output config.pb 
#converting modified_config.json to modified_config.pb
configtxlator proto_encode --input modified_config.json --type common.Config --output modified_config.pb
#finding out the diff and storing in org3_update.pb
configtxlator compute_update --channel_id testcommon --original config.pb --updated modified_config.pb --output org3_update.pb

#need to add an envelope portion to the org3_update.pb
# for that we need to decode org3_update.pb into json add the envelope and encode it into .pb
# decoding into json
configtxlator proto_decode --input org3_update.pb --type common.ConfigUpdate | jq . > org3_update.json
#adding the envelope
echo '{"payload":{"header":{"channel_header":{"channel_id":"testcommon", "type":2}},"data":{"config_update":'$(cat org3_update.json)'}}}' | jq . > org3_update_in_envelope.json
#encoding into pb format
configtxlator proto_encode --input org3_update_in_envelope.json --type common.Envelope --output org3_update_in_envelope.pb


#in cli.Org2 bash (or in the different client from the previous steps)
#
docker exec -e CORE_PEER_LOCALMSPID="Org2MSP" -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/peers/peer0.org2/tls/ca.crt -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/users/Admin@org2/msp -e CORE_PEER_ADDRESS=peer0.org2:7051 -it cli.org2 bash

#signing from org1 cli
peer channel signconfigtx -f org3_update_in_envelope.pb

#send the update in org2 as well which include org2 also 
peer channel update -f org3_update_in_envelope.pb -c testcommon -o orderer.test:7050 —-tls —-cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem 

#bring up the new org
docker-compose -f docker-compose-org3.yaml up -d

#in cli of org1 fetch the genesis block
peer channel fetch 0 testcommon.block -o orderer.test:7050 -c testcommon --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem 

#need to copy the genesis block from org1 cli to org3 cli
#copying to local system
docker cp cli.org1:/opt/gopath/src/github.com/hyperledger/fabric/peer/testcommon.block .
#copying from local to org3
docker cp testcommon.block cli.org3:/opt/gopath/src/github.com/hyperledger/fabric/peer/

#now join org3 to the channel
peer channel join -b testcommon.block
# final check whether the org3 is in sync or not
peer channel getinfo -c testcommon # shud the see the same height as from other peers

#install the chaincode and instantiate and invoke for testing
