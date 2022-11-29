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
```
