#!/bin/bash
parseParameters() {
  para=$1
  echo $para
  shift
  PEER_CONN_PARMS=""
  PEERS=""
  while [ "$#" -gt 0 ]; do
    PEER="peer$1.org$2"
    PEERS="$PEERS $PEER"
    # shift by two to get the next pair of peer/org parameters
    shift
    shift
  done
  # remove leading space for output
  PEERS="$(echo -e "$PEERS" | sed -e 's/^[[:space:]]*//')"
  # echo $PEERS
}


chaincodeGenerateR() {
  parseParameters $@
  echo $PEERS
}

instantiateChaincodeNew() {
  para=$1
  PEER=$2
  ORG=$3
  # setGlobals $PEER $ORG
  # VERSION=${3:-1.0}

  # while 'peer chaincode' command can get the orderer endpoint from the peer
  # (if join was successful), let's supply it directly as we know it using
  # the "-o" option
  # if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
  #   set -x
  #   peer chaincode instantiate -o orderer.example.com:7050 -C $CHANNEL_NAME -n mycc -l ${LANGUAGE} -v ${VERSION} -c '$para' -P "AND ('Org1MSP.peer','Org2MSP.peer')" >&log.txt
  #   res=$?
  #   set +x
  # else
  echo $para
    set -x
    peer chaincode instantiate -o orderer.example.com:7050  -c "${para}" -P "AND ('Org1MSP.peer','Org2MSP.peer')" >&log.txt
    res=$?
    set +x
  # fi
  # cat log.txt
  # verifyResult $res "Chaincode instantiation on peer${PEER}.org${ORG} on channel failed"
  echo "===================== Chaincode is instantiated on peer${PEER}.org${ORG} on channel  ===================== "
  echo
}
