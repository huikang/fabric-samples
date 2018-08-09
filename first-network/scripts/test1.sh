#!/bin/bash

. test2.sh

# chaincodeGenerateR '{"Args":["invoke","a","-100","b","100"]}' 0 1 #space between each parameter

instantiateChaincodeNew '{"Args":["init","a","150","b","0","c","0","d","0"]}' 0 2
