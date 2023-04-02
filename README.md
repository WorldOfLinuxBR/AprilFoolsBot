# AprilFoolsBot

Esse bot foi utilizado na pegadinha de 1º de abril de 2023 realizada no Discord. O objetivo da pegadinha era renomear todos os usuários do servidor adicionando "@Windows 12" em seus nomes de usuário.

O código fonte do bot é escrito em Go e usa a biblioteca "discordgo" para se conectar ao servidor do Discord. O bot também usa o MongoDB para armazenar os nomes de usuário dos membros do servidor antes da pegadinha.


## Como o bot funciona

O bot tem três comandos que podem ser usados apenas pelo proprietário do bot:

    `!aprilfools`: Esse comando é usado para renomear todos os usuários do servidor adicionando "@Windows 12" em seus nomes de usuário. O bot obtém uma lista de todos os membros do servidor. Em seguida, o bot atualiza o apelido de cada membro do servidor para adicionar "@Windows 12". O bot também verifica se o nome de usuário do membro já contém "@Windows 12" para evitar duplicatas. O nome original e o novo nome do membro são registrados no terminal.

    `!backupUsernames`: Esse comando foi usado para fazer backup dos nomes de usuário dos membros do servidor antes da pegadinha. O bot obtém uma lista de todos os membros do servidor e salva o nome de usuário de cada membro no MongoDB usando a função coll.InsertOne. O nome de usuário é salvo no MongoDB como um documento que contém o ID do usuário e o nome de usuário. O comando também deleta todos os documentos anteriores da coleção antes de fazer o backup.
    `!undoAprilFools`: Esse comando (infelizmente) foi usado para renomear todos os usuários para o nick anterior a pegadinha, utilizando os dados do MongoDB.

O bot usa variáveis de ambiente para armazenar informações sensíveis, como o token do Discord, o URI do MongoDB, o ID do servidor e o ID do proprietário do bot. Essas informações são carregadas do arquivo `.env` no início da execução do programa.

## Como executar o bot

Para executar o bot, é necessário criar um bot no Discord e obter o token do bot. É necessário também criar um arquivo `.env` na raiz do projeto com as seguintes variáveis:

```bash
DISCORD_TOKEN=<seu token do Discord>
MONGO_URI=<sua URI do MongoDB>
GUILD_ID=<o ID do seu servidor Discord>
OWNER_ID=<seu ID de usuário do Discord>
```


Certifique-se de ter o Go instalado na sua máquina antes de prosseguir. Em seguida, abra um terminal na pasta raiz do projeto e execute os seguintes comandos:

```
go mod tidy
go run .
```

> O bot deve estar online e pronto para responder aos comandos. Lembre-se de que esses comandos só podem ser usados pelo proprietário do bot.
