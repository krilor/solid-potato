GET {{host}}/transactions
[Query]
status: APPROVED
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$" count == 75
jsonpath "$[?(@.status == 'APPROVED')]" count == 75

GET {{host}}/transactions
[Query]
status: REJECTED
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$" count == 12
jsonpath "$[?(@.status == 'REJECTED')]" count == 12

GET {{host}}/transactions
[Query]
status: PENDING
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$" count == 11
jsonpath "$[?(@.status == 'PENDING')]" count == 11

GET {{host}}/transactions
[Query]
status: INVALID
HTTP 400
