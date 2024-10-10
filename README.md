Para rodar o projeto basta estar dentro do diretório do repositório e rodar o seguinte comando:
**sudo go run src/main.go
**
Se quiser, é possível especificar a porta de cliente e servidor usando as seguintes flags:
**--server-port
--client-port
**
Ou seja:
**sudo go run src/main.go --server-port <port> --client-port <port>**

Máquinas necessitam ter o ebpftool instalado.
