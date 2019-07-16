
setPrefereces:
______________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  -C preferenceschannel -n preferences --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":["sp","{\"msisdn\":\"8848022338\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crmno\":\"123456\",\"srvac\":\"1\",\"ptype\":\"2\",\"sts\":\"A\"}"]}'

	OutPut On Success:
		"{\"message\":\"Add Preferences Success\",\"msisdn\":\"8848022338\",\"trxnid\":\"2d7b7f1f7bfbe6766d3db35398be7532bb0abaab66d938df92dd0dbd30b9a2c0\"}"


portOut:
_______
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["po","{\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"lrn\":\"1234\",\"uts\":\"1557311911\",\"srvac\":\"2\"}"]}
	OutPut On Success:
		"{\"message\":\"Portout is Success\",\"msisdn\":\"8848022338\",\"trxnid\":\"407504a015683eac5b6a1dd5d5d55a1241519ed23f788144e9fe3e24cb24d39c\"}"


deletePreferences:
_________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["dp","{\"msisdn\":\"7702906226\",\"uts\":\"1557233449\"}"]}'

	OutPut On Success:
		"{\"message\":\"Delete Preferences Success\",\"msisdn\":\"7702906226\",\"trxnid\":\"e080b713281796aa71ae9657b75831c9a7a4e6ade9cc16a21c1d7534b27a1f59\"}"


batchPreferences:
_________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["abp","{\"msisdn\":\"8848022331\",\"svcprv\":\"VI\",\"reqno\":\"8848022331\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"uts\":\"1557314556\",\"cts\":\"1557314556\",\"crmno\":\"9848022339\",\"srvac\":\"1\",\"ptype\":\"2\",\"sts\":\"A\"}","{\"msisdn\":\"8848022332\",\"svcprv\":\"VI\",\"reqno\":\"8848022332\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"uts\":\"1557314557\",\"cts\":\"1557314556\",\"crmno\":\"8848022337\",\"srvac\":\"3\",\"ptype\":\"2\",\"sts\":\"A\"}"]}'
	
	OutPut On Success:
		"{\"message\":\"Batch Preferences Success\",\"msisdn_f\":null,\"trxnid\":\"d0335a343438d82ee70af709475a0e05c449d686b34da9b4fbbb1930fecbe1dc\"}"


batchPortOut:
____________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["bpo","{\"msisdn\":\"8848022331\",\"svcprv\":\"JI\",\"lrn\":\"1234\",\"uts\":\"1557315063\",\"srvac\":\"2\"}","{\"msisdn\":\"8848022332\",\"svcprv\":\"BL\",\"lrn\":\"1234\",\"uts\":\"1557315063\",\"srvac\":\"2\"}"]}'
	OutPut On Success:
		"{\"message\":\"Batch PortOut Success\",\"msisdn_f\":null,\"trxnid\":\"b7eced25679781115352ea0e342a5ace7ba2b9e642ffb25b7e2b5704cb7617cc\"}"

batchDeletePreferences:
______________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["dbp","{\"msisdn\":\"8848022334\",\"uts\":\"1557314556\"}","{\"msisdn\":\"8848022333\",\"uts\":\"1557314556\"}"]}' 
	OutPut On Success:
		"{\"message\":\"Batch Delete Success\",\"msisdn_f\":null,\"trxnid\":\"4f7d8375bf6a6f2b9982141b4ee2e74ff7b08d06db46118cf46e1b07c98f9b1c\"}"

snapBackChurn:
_____________
	Input:
		    peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["sbc","{\"msisdn\":\"8748022338\",\"svcprv\":\"AI\",\"lrn\":\"1234\",\"uts\":\"1557311911\",\"srvac\":\"2\"}"]}
        OutPut On Success:
                "{\"message\":\"SnapBackChurn is Success\",\"msisdn\":\"8748022338\",\"trxnid\":\"407504a015673eac5b6a1dd5d5d55a1241519ed23f788144e9fe3e24cb24d39c\"}"

batchSnapBackChurn:
___________________
	Input:
	       peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["bsbc","{\"msisdn\":\"8848022335\",\"svcprv\":\"JI\",\"lrn\":\"1234\",\"uts\":\"1557315063\",\"srvac\":\"2\"}","{\"msisdn\":\"8848022336\",\"svcprv\":\"BL\",\"lrn\":\"1234\",\"uts\":\"1557315063\",\"srvac\":\"2\"}"]}'
        OutPut On Success:
                "{\"message\":\"Batch SnapBackChurn Success\",\"msisdn_f\":null,\"trxnid\":\"b7eced25679781125352ea0e342a5ace7ba2b9e642ffb25b7e2b5704cb7617cc\"}"

	OutPut On Success:
	

getPreferencesByMsisdn:
______________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["pd","8848022338"]}'
	
	OutPut On Success:
		"{\"preferences\":{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"2\",\"ptype\":\"2\"},\"status\":\"true\"}"


queryPreferences:
________________
	Input:
	______
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qp","{\"selector\":{\"msisdn\":\"8848022338\"}}"]}'

	OutPut On Success:
		"{\"preferences:\":[{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"2\",\"ptype\":\"2\"}],\"status\":\"true\"}"

	Input2:
	_______
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qp","{\"selector\":{\"msisdn\":\"8848022330\"}}"]}'
	OutPutOnSuccess:
		"{\"preferences:\":null,\"status\":\"true\"}"


	Input3:
	_______
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qp","{\"selector\":{\"obj\":\"Preferences\"}}"]}'
	
	OutPutOnSuccess:
		"{\"preferences:\":[{\"obj\":\"Preferences\",\"msisdn\":\"1008238798\",\"svcprv\":\"VI\",\"reqno\":\"567588998888888888\",\"rmode\":\"0\",\"ctgr\":\"\",\"cmode\":\"10\",\"day\":\"31,32,33,35\",\"time\":\"21,22,23,24,25,28,29\",\"lrn\":\"1234\",\"cts\":\"1558001484\",\"uts\":\"1558077982\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702111116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702121112\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702121116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702901116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"1560709099\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702902116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"1560709550\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702903116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"1560709719\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702906116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"1560709600\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702906226\",\"svcprv\":\"VI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"2234\",\"cts\":\"1557233447\",\"uts\":\"1557233449\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"123456\",\"sts\":\"T\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702906227\",\"svcprv\":\"VI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"2234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702911116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"8848022331\",\"svcprv\":\"JI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557315063\",\"crtr\":\"org1.example.com\",\"uby\":\"jio.com\",\"crmno\":\"9848022339\",\"sts\":\"A\",\"srvac\":\"2\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"8848022332\",\"svcprv\":\"BL\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557315063\",\"crtr\":\"org1.example.com\",\"uby\":\"bsnl.com\",\"crmno\":\"8848022337\",\"sts\":\"A\",\"srvac\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"8848022333\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557314556\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"8848022333\",\"sts\":\"T\",\"srvac\":\"3\"},{\"obj\":\"Preferences\",\"msisdn\":\"8848022334\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557314556\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022339\",\"sts\":\"T\",\"srvac\":\"1\"},{\"obj\":\"Preferences\",\"msisdn\":\"9533689255\",\"svcprv\":\"ID\",\"reqno\":\"110215580743696240\",\"rmode\":\"0\",\"ctgr\":\"0\",\"cmode\":\"10\",\"day\":\"30\",\"time\":\"20\",\"lrn\":\"1700\",\"cts\":\"1558069079\",\"uts\":\"1558077930\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9533689266\",\"svcprv\":\"ID\",\"reqno\":\"110315580106338932\",\"rmode\":\"0\",\"ctgr\":\"0\",\"cmode\":\"10\",\"day\":\"30\",\"time\":\"20\",\"lrn\":\"1700\",\"cts\":\"1558001934\",\"uts\":\"1558010636\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9652693062\",\"svcprv\":\"ID\",\"reqno\":\"110315578147738425\",\"rmode\":\"0\",\"ctgr\":\"2,3,4\",\"cmode\":\"12\",\"day\":\"31,33\",\"time\":\"21,22,23,27,29\",\"lrn\":\"1700\",\"cts\":\"1557999735\",\"uts\":\"1557999920\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9652693962\",\"svcprv\":\"VO\",\"reqno\":\"110315579999783364\",\"rmode\":\"2\",\"ctgr\":\"2,3,4,5\",\"cmode\":\"12,13\",\"day\":\"31,33,35\",\"time\":\"23,25,27,29\",\"lrn\":\"1700\",\"cts\":\"1557999938\",\"uts\":\"1557999978\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022330\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557928214\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022337\",\"sts\":\"T\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022331\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557928214\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022339\",\"sts\":\"T\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022337\",\"svcprv\":\"BL\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557315063\",\"crtr\":\"org1.example.com\",\"uby\":\"bsnl.com\",\"crmno\":\"9848022337\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022338\",\"svcprv\":\"AI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22,23\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"1234567\",\"sts\":\"A\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022339\",\"svcprv\":\"JI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557315063\",\"crtr\":\"org1.example.com\",\"uby\":\"jio.com\",\"crmno\":\"9848022339\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"}],\"status\":\"true\"}"

	Using Indexes: 
		Input4 SearchBy ServiceProvider:
			 peer chaincode invoke -o orderer.example.com:7050  -C preferenceschannel -n preferences --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":["qp","{\"selector\":{\"svcprv\":\"VI\"},\"use_index\":\"preferencesSearchBySvcprv\"}"]}'
		
	OutPut On Success:
			"{\"preferences:\":[{\"obj\":\"Preferences\",\"msisdn\":\"8848022331\",\"svcprv\":\"VI\",\"reqno\":\"8848022331\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557314556\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022339\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"8848022332\",\"svcprv\":\"VI\",\"reqno\":\"8848022332\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557314557\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"8848022337\",\"sts\":\"A\",\"srvac\":\"3\",\"ptype\":\"2\"}],\"status\":\"true\"}"

		Input5 SearchBy ReqNumber:
			 peer chaincode invoke -o orderer.example.com:7050  -C preferenceschannel -n preferences --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":["qp","{\"selector\":{\"reqno\":\"12345678\"},\"use_index\":\"preferencesSearchByReqno\"}"]}'

		OutPut On Success:			
			"{\"preferences:\":[{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"}],\"status\":\"true\"}"


		Input6 SearchBy UpdateTs:
			peer chaincode invoke -o orderer.example.com:7050  -C preferenceschannel -n preferences --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":["qp","{\"selector\":{\"uts\":\"1557233447\"},\"use_index\":\"preferencesSearchByUts\"}"]}'

		OutPut On Success:
		 	"{\"preferences:\":[{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"}],\"status\":\"true\"}"

		Input SearchBy Sts:
			peer chaincode invoke -o orderer.example.com:7050  -C preferenceschannel -n preferences --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -c '{"Args":["qp","{\"selector\":{\"sts\":\"T\"},\"use_index\":\"preferencesSearchBySts\"}"]}'

		OutPut On Success:
			"{\"preferences:\":[{\"obj\":\"Preferences\",\"msisdn\":\"9848022330\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557928214\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022337\",\"sts\":\"T\",\"srvac\":\"\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022331\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"\",\"cmode\":\"\",\"day\":\"\",\"time\":\"\",\"lrn\":\"1234\",\"cts\":\"1557314556\",\"uts\":\"1557928214\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"9848022339\",\"sts\":\"T\",\"srvac\":\"\"}],\"status\":\"true\"}"

historyPreferences:
___________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["hp","8848022338"]}'	

	OutPut On Success:
		"{\"preferences\":[{\"obj\":\"Preferences\",\"msisdn\":\"8848022338\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"VI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"2234\",\"cts\":\"1557233447\",\"uts\":\"1557233447\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"2\",\"ptype\":\"2\"}],\"status\":\"true\"}"

	Input2:
		 peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["hp","984802233"]}'

	OutPutOnSuccess:
		"{\"preferences\":null,\"status\":\"true\"}"



queryPreferencesWithPagination:
______________________________
	Input:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qpp","{\"selector\":{\"obj\":\"Preferences\"}}","2",""]}'
	
	OutPutOn Success:
		"{\"bookmark\":\"g1AAAABEeJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYpzmZsbGBmCgBlIBQdMBZpcFgBMaRC_\",\"preferences\":[{\"obj\":\"Preferences\",\"msisdn\":\"1008238798\",\"svcprv\":\"VI\",\"reqno\":\"567588998888888888\",\"rmode\":\"0\",\"ctgr\":\"\",\"cmode\":\"10\",\"day\":\"31,32,33,35\",\"time\":\"21,22,23,24,25,28,29\",\"lrn\":\"1234\",\"cts\":\"1558001484\",\"uts\":\"1558077982\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702111116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"}],\"recordscount\":2,\"status\":\"true\"}"

	Input2:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qpp","{\"selector\":{\"obj\":\"Preferences\"}}","2","g1AAAABEeJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYpzmZsbGBmCgBlIBQdMBZpcFgBMaRC_"]}'	

	OutPutOnSuccess:
		"{\"bookmark\":\"g1AAAABEeJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYpzmZsbGBkaGRoamoFUcMBUoMllAQBMkRDB\",\"preferences\":[{\"obj\":\"Preferences\",\"msisdn\":\"7702121112\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"7702121116\",\"svcprv\":\"VI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1\",\"cmode\":\"1,2\",\"day\":\"1\",\"time\":\"123\",\"lrn\":\"1234\",\"cts\":\"11111\",\"uts\":\"123456\",\"crtr\":\"org1.example.com\",\"uby\":\"org1.example.com\",\"crmno\":\"\",\"sts\":\"A\",\"srvac\":\"1\",\"ptype\":\"2\"}],\"recordscount\":2,\"status\":\"true\"}"

	Input3:
		peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qpp","{\"selector\":{\"svcprv\":\"AI\"}}","2",""]}'
	
	OutPutOnSuccess:
		"{\"bookmark\":\"g1AAAABEeJzLYWBgYMpgSmHgKy5JLCrJTq2MT8lPzkzJBYpzWVqYWBgYGRkbW4BUcMBUoMllAQBPrRDn\",\"preferences\":[{\"obj\":\"\",\"msisdn\":\"8848022338\",\"svcprv\":\"AI\",\"reqno\":\"123456789\",\"rmode\":\"1\",\"ctgr\":\"1,4,5\",\"cmode\":\"11\",\"day\":\"21,22\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"123456\",\"sts\":\"A\",\"srvac\":\"2\",\"ptype\":\"2\"},{\"obj\":\"Preferences\",\"msisdn\":\"9848022338\",\"svcprv\":\"AI\",\"reqno\":\"12345678\",\"rmode\":\"1\",\"ctgr\":\"1,2,3,4\",\"cmode\":\"10\",\"day\":\"21,22,23\",\"time\":\"31,32\",\"lrn\":\"1234\",\"cts\":\"1557233447\",\"uts\":\"1557311911\",\"crtr\":\"org1.example.com\",\"uby\":\"airtel.com\",\"crmno\":\"1234567\",\"sts\":\"A\",\"srvac\":\"1\",\"2\"}],\"recordscount\":2,\"status\":\"true\"}"

	Input4:
		 peer chaincode invoke -o orderer.example.com:7050  --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C preferenceschannel -n preferences -c '{"Args":["qpp","{\"selector\":{\"svcprv\":\"unknown\"}}","2",""]}'
	
	OutPutOnSuccess:
		"{\"bookmark\":\"nil\",\"preferences\":null,\"recordscount\":0,\"status\":\"true\"}"


