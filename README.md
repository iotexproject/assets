assets
======

# Deploy

```
export KEY=xxxx
export SITE_URL=http://localhost:3000
```

# APIs

## Token list

```
curl -i http://localhost:3000/tokenlist/iotex
```

## Token detials

```
curl -i http://localhost:3000/token/iotex/0xb8403ffba4d0af0e430b128c5569e335ec00c4c9
```

## Token image for NFTs

```
curl -i http://localhost:3000/token/iotex/0xb8403ffba4d0af0e430b128c5569e335ec00c4c9/image/1
curl -i http://localhost:3000/token/iotex/0x30582ede7fadeba4973dd71f1ce157b7203171ea/image/1
curl -i http://localhost:3000/token/ethereum/0x306b1ea3ecdf94ab739f1910bbda052ed4a9f949/image/1
curl -i http://localhost:3000/token/ethereum/0xb668beb1fa440f6cf2da0399f8c28cab993bdd65/image/1
curl -i http://localhost:3000/token/ethereum/0x57f1887a8bf19b14fc0df6fd9b2acc9af147ea85/image/53759650996537692076129934293629512578081917330486191194657099486799331644576
curl -i http://localhost:3000/token/ethereum/0x90350ac498e9e78943f1286053c1985efea0561a/image/1
curl -i http://localhost:3000/token/ethereum/0xdc85866ddd95fa9b7c856944fab128902ca8c60f/image/1
curl -i http://localhost:3000/token/ethereum/0x33fd426905f149f8376e227d0c9d3340aad17af1/image/1
curl -i http://localhost:3000/token/ethereum/0xb47e3cd837ddf8e4c57f05d70ab865de6e193bbb/image/1
// TODO process error code
curl -i http://localhost:3000/token/ethereum/0x82c7a8f707110f5fbb16184a5933e9f78a34c6ab/image/5986037213260671
curl -i http://localhost:3000/token/ethereum/0x394e3d3044fc89fcdd966d3cb35ac0b32b0cda91/image/1
```

## Own Tokens

```
curl -i http://localhost:3000/account/1/own/0x0000000000000000000000000000000000000001?skip=0&first=10
curl -i http://localhost:3000/account/4689/own/0x4309b22dfe0d062f54763e1a9aec74a636fa3276?skip=0&first=10
curl -i http://localhost:3000/account/4690/own/0xfeed8588af2ba5d18499d97a53de9e2d504d3641?skip=0&first=10
```

# Docker

## Build

```
docker build -t assets:latest .
```

## Run

```
docker run -d -p 3000:3000 \
  -e KEY=XXXXX \
  -e SITE_URL=https://nft.iopay.me \
  assets
```
