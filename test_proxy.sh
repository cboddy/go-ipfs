#!/bin/bash

#
# serve the thing
#
SERVE_PATH=$(mktemp -d)
echo "YOU ARE THE CHAMPION MY FRIEND" > $SERVE_PATH/index.txt
cd $SERVE_PATH
# serve this on port 8000
python -m SimpleHTTPServer 8000 & 

cd -

IPFS=cmd/ipfs/ipfs

PATH1=$(mktemp -d)
PATH2=$(mktemp -d)

RECEIVER_LOG=$PATH1/log.log
SENDER_LOG=$PATH2/log.log

export IPFS_PATH=$PATH1
$IPFS init >> $RECEIVER_LOG 2>&1
$IPFS config --json Experimental.Libp2pStreamMounting true >> $RECEIVER_LOG 2>&1
$IPFS config --json Addresses.API "\"/ip4/127.0.0.1/tcp/6001\"" >> $RECEIVER_LOG 2>&1
$IPFS config --json Addresses.Gateway "\"/ip4/127.0.0.1/tcp/8081\"" >> $RECEIVER_LOG 2>&1
$IPFS config --json Addresses.Swarm "[\"/ip4/0.0.0.0/tcp/7001\", \"/ip6/::/tcp/7001\"]" >> $RECEIVER_LOG 2>&1
$IPFS daemon >> $RECEIVER_LOG 2>&1 &
# wait for daemon to start.. maybe?
# ipfs id returns empty string ifwe don't wait here..
sleep 5 
$IPFS p2p listener open test /ip4/127.0.0.1/tcp/8000 >> $RECEIVER_LOG 2>&1
FIRST_ID=$($IPFS id -f "<id>")

export IPFS_PATH=$PATH2
$IPFS init >> $SENDER_LOG 2>&1
$IPFS config --json Experimental.Libp2pStreamMounting true >> $SENDER_LOG 2>&1 
$IPFS daemon >> $SENDER_LOG 2>&1 &
# wait for daemon to start.. maybe?
sleep 5 



# request to SENDER that should go via proxy 
curl http://localhost:5001/proxy/http/$FIRST_ID/test/index.txt



echo "******************" 
echo link http://localhost:5001/proxy/http/$FIRST_ID/test/index.txt
echo "******************" 
echo "RECEIVER IPFS LOG " $RECEIVER_LOG 
echo "******************" 
cat $RECEIVER_LOG

echo "******************" 
echo "SENDER IPFS LOG " $SENDER_LOG
echo "******************" 
cat $SENDER_LOG


