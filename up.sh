docker exec cli.org1 bash -c 'peer channel create -c testcommon -f ./channels/testcommon.tx -o orderer.test:7050 -t 60s --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem'

docker exec cli.org1 bash -c 'peer channel join -o orderer.test:7050 -b testcommon.block --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem'
docker exec cli.org2 bash -c 'peer channel join -o orderer.test:7050 -b testcommon.block --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem'

docker exec cli.org1 bash -c 'peer channel update -o orderer.test:7050 -c testcommon -f ./channels/org1-testcommon-anchor.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem'
docker exec cli.org2 bash -c 'peer channel update -o orderer.test:7050 -c testcommon -f ./channels/org2-testcommon-anchor.tx --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/test/orderers/orderer.test/msp/tlscacerts/tlsca.test-cert.pem'


#docker exec cli.org1 bash -c 'peer chaincode install -o orderer.test:7050 -p simplyfi/simplyfi/training -n training -v 1.0'
#docker exec cli.org2 bash -c 'peer chaincode install -o orderer.test:7050 -p simplyfi/simplyfi/training -n training -v 1.0'

#docker exec cli.org1 bash -c "peer chaincode instantiate -o orderer.test:7050 -C testcommon -n training -v 1.0 -c '{\"Args\":[]}'"
#docker exec cli.org1 bash -c "peer chaincode upgrade -o orderer.telco.com:7050 -C testcommon -n training -v 1.1 -c '{\"Args\":[]}'"
