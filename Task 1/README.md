## Проектное задание номер 1 – Александр Щепетов
---
### Развертывание сети + деплой чейнкода
  ```
  ./network/network.sh up createChannel -c mychannel -ca -verbose && ./network/network.sh deployCC -ccn census -ccp ../chaincode/chaincode-go -ccl go -cci InitLedger
  ```

### Консольное приложение
  ```
  cd application && go run cli.go --help
  ```