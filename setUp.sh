sudo rm -r orderer channels
mkdir orderer channels
#cryptogen generate --config crypto-config.yaml
export FABRIC_CFG_PATH=./ #path to the configtx.yaml

configtxgen -profile Genesis -outputBlock ./orderer/genesis.block

configtxgen -profile CommonChannel -outputCreateChannelTx ./channels/testcommon.tx -channelID testcommon
configtxgen -profile CommonChannel -outputAnchorPeersUpdate ./channels/org1-testcommon-anchor.tx -channelID testcommon -asOrg Org1MSP
configtxgen -profile CommonChannel -outputAnchorPeersUpdate ./channels/org2-testcommon-anchor.tx -channelID testcommon -asOrg Org2MSP
