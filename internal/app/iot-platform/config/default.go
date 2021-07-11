package config

// nolint:lll
const defaultConfig = `
central-server:
  address: :65432
  read-timeout: 2m
  write-timeout: 2m
  graceful-timeout: 5s
local-server:
  address: :54321
  read-timeout: 2m
  write-timeout: 2m
  graceful-timeout: 5s
cooler:
  type: 2
  address: :6542
  read-timeout: 2m
  write-timeout: 2m
  graceful-timeout: 5s
light-bulb:
  type: 3
  address: :6543
  read-timeout: 2m
  write-timeout: 2m
  graceful-timeout: 5s
temperature:
  type: 0
  local-server-address: http://localhost:54321
  read-timeout: 2m
light:
  type: 1
  local-server-address: http://localhost:54321
  read-timeout: 2m
jwt:
  expiration: 5m
  secret: 'SECRET_TOKEN'
`
