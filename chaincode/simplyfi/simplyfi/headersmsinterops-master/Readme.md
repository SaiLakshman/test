<!-- Register Header -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["rh","{\"hid\":\"QT0411111111111\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE41\",\"ctgr\":\"8\",\"cts\":\"1234567878\",\"uts\":\"456787678\"}"]}'

<!-- Register Bulk Header -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["rbh","{\"hid\":\"QT041111111111111102\",\"peid\":\"A11111111101\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE2\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}","{\"hid\":\"QT041111111111111103\",\"peid\":\"A11111111102\",\"htyp\":\"T\",\"cli\":\"BLOCKCUBE3\",\"ctgr\":\"8\",\"cts\":\"456789\",\"uts\":\"456787678\"}"]}'

<!-- Query Header with Multiple cli's  -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["qh","BLOCKCUBE","BLOCKCUBE3"]}'

<!-- Update Header Status -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["uhs","{\"cli\":\"BLOCKCUBE\",\"sts\":\"I\",\"uts\":\"2345678\"}"]}'

<!-- queryHistory by Parameters -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["qhbp","{\"typ\":\"hid\",\"hid\":\"QT041111111111111101\"}"]}'

<!-- History of a header -->
peer chaincode invoke -o orderer-simplyfi-co-in:7050 -n header -C telcocommon  -c '{"args":["hfh","{\"cli\":\"sdaf\"}"]}'

<!-- Query History with Pagination -->
 peer chaincode invoke -C telcocommon -o orderer-simplyfi-co-in:7050 -n header -c '{"args": ["qhwp","{\"selector\":{\"htyp\":\"T\"}}","4",""]}'
