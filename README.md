# Shapeshifter-trs 

`docker build -t qredo/shapeshifter-trs --build-arg token=yourGitHubAuthToken .`

# Add this to your ~/.bash_profile

## Mac

`export DYLD_LIBRARY_PATH=/usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib:$DYLD_LIBRARY_PATH`

## Ubuntu

`export LD_LIBRARY_PATH=/usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib:$LD_LIBRARY_PATH`
