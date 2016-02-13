ESXCloud Go Cli
===============

The repo for the ESXCloud go cli.

## Setup

The project requires go version 1.5+. You can download and install go from: https://golang.org/dl/

Decide a folder as the GOPATH, e.g. ~/go.
    
    1. mkdir -p ~/go/src/github.com/vmware/
    2. cd ~/go/src/github.com/vmware/ 
    3. git clone (this repo from gerrit or github)
    4. export GOPATH=~/go
    5. export PATH=$PATH:~/go/bin
    6. make tools
    7. godep restore 


## Usage

To run the test:
    
      make test
      
To build the excecutables:
    
      make build
      
The executables are generated under photon-controller-cli/bin folder.
    
To run and verify the CLI:
    
       ./bin/photon -v
        
    
## Pick up Changes from SDK

When there are changes in SDK, wait for them promoted to **_master_** branch on **_github.com_**.

Follow the steps below:
    
    1. go get -u github.com/vmware/photon-controller-go-sdk/photon
    2. godep update github.com/vmware/photon-controller-go-sdk/photon
    
Before comitting the change, carefully inspect the changes to Godeps, for example with git diff or SourceTree.
    
Then you can commit and submit the change.
