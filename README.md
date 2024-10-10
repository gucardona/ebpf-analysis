#Requisitos
Go: Certifique-se de ter o Go instalado em sua máquina. Para instalar o Go, siga as instruções no [site oficial do Go](https://golang.org/doc/install).
ebpftool: As máquinas necessitam ter o ebpftool instalado. Você pode instalar o ebpftool seguindo as instruções fornecidas [neste repositório](https://github.com/sysrepo/ebpftool).


#Instalação
##Clone o repositório e navegue até o diretório do projeto:
git clone https://github.com/gucardona/ebpf-analysis.git

#Executando o Projeto
##Para rodar o projeto, siga os passos abaixo:

##Entre no diretório do repositório:
cd ebpf-analysis

##Execute o comando go:
sudo go run src/main.go

##Especificar Portas de Cliente e Servidor (Opcional)
##Se desejar, você pode especificar as portas de cliente e servidor usando as seguintes flags:

--server-port para definir a porta do servidor
--client-port para definir a porta do cliente

##Exemplo de uso:
sudo go run src/main.go --server-port <PORTA_SERVIDOR> --client-port <PORTA_CLIENTE>
Substitua <PORTA_SERVIDOR> e <PORTA_CLIENTE> pelas portas que deseja usar para o servidor e o cliente, respectivamente.
