# My own DNS UDP Server

A DNS (Domain Name System) server is a system that translates human-readable domain names, into IP addresses, which are used by devices to identify each other on a network. DNS servers allow devices connected to the internet to find and communicate with each other using domain names, without needing to remember numerical IP addresses.

### Key Concepts:

1. **UDP Protocol (User Datagram Protocol):**
   The DNS server in this implementation uses UDP, a communication protocol that is faster than others like TCP but does not guarantee the delivery or order of data. Despite this, UDP is commonly used in DNS because queries are typically small and do not require a persistent connection.

2. **Name Resolution (DNS Query):**
   DNS servers operate through queries. When a user wants to access a website, for example, their device sends a DNS query to resolve the domain name into an IP address

3. **DNS Recursion:**
   Some DNS servers act as "recursive resolvers." This means that they not only respond to queries for domain names they know but also search for answers from other DNS servers if they do not have the information. Recursive resolvers are very common on the internet.

4. **Responding to a Query:**
   After receiving a query, the DNS server sends back the corresponding information, usually the IP address associated with the requested domain name. If it cannot resolve the domain name, the server may return an error indicating that the address could not be found.

5. **DNS Records:**
   DNS responses can include different types of records, such as:
   - **A** (IPv4 address)
   - **AAAA** (IPv6 address)
   - **CNAME** (canonical name, used for aliases)
   - **MX** (mail server)
   - And others.

6. **DNS Packet Header:**
   A DNS packet contains a header with information such as the query identifier, the number of questions and answers, and error codes. This header helps the server and the client manage and interpret responses.

7. **Name Compression:**
   To make data transfer more efficient, DNS packets can use a mechanism called name compression, which allows repeated domain names in a query or response to be represented compactly.

### General Operation of a UDP DNS Server:

1. **Listening for Queries:**
   The DNS server listens for incoming queries over the UDP protocol on a specific port (in this case, port 2053).

2. **Processing the Query:**
   Upon receiving a query, the server analyzes the requested domain name and decides whether it has the answer or needs to obtain it from another DNS server (if it's acting as a recursive resolver).

3. **Generating the Response:**
   If the server has the information, it generates a response with the relevant DNS records (such as the IP address). If it doesn't have the answer and is configured to resolve queries recursively, it forwards the query to other DNS servers.

4. **Responding to the Client:**
   Finally, the server responds to the client with the requested records or with an error message if it cannot resolve the domain.

In summary, this type of DNS UDP server acts as a translator for domain names to IP addresses on a network and is essential for enabling effective communication between devices on the internet.

## How to run

```bash
cd cmd
go run .
```

## How to test in another terminal

```bash
cd test
go run .
```
