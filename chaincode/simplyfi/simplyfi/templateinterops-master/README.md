# Templateinterops

Consent and Content Template chaincode

#setTemplate
peer chaincode invoke -C telcocommon -n templates -c '{"args":["st","{\"urn\": \"120756782782\",\"peid\": \"120145678912\",\"cli\": [\"Header1\", \"Header2\"],\"tname\": \"Template1\",\"ctyp\": \"P\",\"ctgr\": \"1\",\"tcont\": \"Dear Subscriber, This is Template Content\",\"cts\": \"1511478902\",\"uts\": \"1511478902\",\"sts\": \"A\",\"csty\":\"1\",\"coty\":\"U\",\"vars\":\"1\",\"tmid\":\"12345\",\"ttyp\":\"CT\"}"]}'
#queryTemplate
peer chaincode invoke -C telcocommon -n templates -c '{"args": ["qt","{\"selector\":{\"urn\":\"120756782782\"}}"]}'
#set batch templates
peer chaincode invoke -C telcocommon -n templates -c '{"args":["abt","{\"urn\": \"120756782782\",\"peid\": \"120145678912\",\"cli\": [\"Header1\", \"Header2\"],\"tname\": \"Template1\",\"ctyp\": \"P\",\"ctgr\": \"1\",\"tcont\": \"Dear Subscriber, This is Template Content\",\"cts\": \"1511478902\",\"uts\": \"1511478902\",\"sts\": \"A\",\"csty\":\"1\",\"coty\":\"U\",\"vars\":\"1\",\"tmid\":\"12345\",\"ttyp\":\"CT\"}","{\"urn\": \"120756782783\",\"peid\": \"120145678912\",\"cli\": [\"Header1\", \"Header2\"],\"tname\": \"Template1\",\"ctyp\": \"P\",\"ctgr\": \"1\",\"tcont\": \"Dear Subscriber, This is Template Content\",\"cts\": \"1511478902\",\"uts\": \"1511478902\",\"sts\": \"A\",\"csty\":\"1\",\"coty\":\"U\",\"vars\":\"1\",\"tmid\":\"12345\",\"ttyp\":\"CS\"}"]}'
#template history
peer chaincode invoke -C telcocommon -n templates -c '{"args": ["th","120756782782"]}'
#update template status
peer chaincode invoke -C telcocommon -n templates -c '{"args": ["uts","120756782782","A","1511478902"]}'
#query template with pagination
peer chaincode invoke -C telcocommon -n templates -c '{"args": ["qtp","{\"selector\":{\"urn\":\"120756782782\"}}","3",""]}'
