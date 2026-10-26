# Expected counts are based on the data loaded from case/transaksjoner.csv
# (after de-duplication by reference). 98 transactions in total, between
# 2024-05-30 and 2025-05-26.
# "from" is inclusive and "to" is exclusive.

# No time filter returns everything
GET {{host}}/transactions
HTTP 200
[Asserts]
jsonpath "$" count == 98

# from and to combined
GET {{host}}/transactions
[Query]
from: 2025-01-01T00:00:00Z
to: 2025-02-01T00:00:00Z
HTTP 200
[Asserts]
jsonpath "$" count == 6

# from only
GET {{host}}/transactions
[Query]
from: 2025-01-01T00:00:00Z
HTTP 200
[Asserts]
jsonpath "$" count == 36

# to only. Together with the from-only request above this covers all transactions.
GET {{host}}/transactions
[Query]
to: 2025-01-01T00:00:00Z
HTTP 200
[Asserts]
jsonpath "$" count == 62

# from is inclusive: the transaction at exactly 2024-06-01T13:16:29+02:00 is included
GET {{host}}/transactions
[Query]
from: 2024-06-01T11:16:29Z
to: 2024-06-01T11:16:30Z
HTTP 200
[Asserts]
jsonpath "$" count == 1
jsonpath "$[0].reference" == "5f00d449-7d34-4147-9486-b8b2cfa52caa"

# to is exclusive: the same transaction is excluded when it is the upper bound
GET {{host}}/transactions
[Query]
from: 2024-06-01T00:00:00Z
to: 2024-06-01T11:16:29Z
HTTP 200
[Asserts]
jsonpath "$" count == 0

# Offsets are respected
GET {{host}}/transactions
[Query]
from: 2024-06-01T13:16:29+02:00
to: 2024-06-01T13:16:30+02:00
HTTP 200
[Asserts]
jsonpath "$" count == 1
jsonpath "$[0].reference" == "5f00d449-7d34-4147-9486-b8b2cfa52caa"

# from after to gives an empty result
GET {{host}}/transactions
[Query]
from: 2025-02-01T00:00:00Z
to: 2025-01-01T00:00:00Z
HTTP 200
[Asserts]
jsonpath "$" count == 0

# Time filters combine with the status filter
GET {{host}}/transactions
[Query]
from: 2025-01-01T00:00:00Z
to: 2025-02-01T00:00:00Z
status: APPROVED
HTTP 200
[Asserts]
jsonpath "$" count == 4
jsonpath "$[?(@.status == 'APPROVED')]" count == 4

# Invalid values are rejected
GET {{host}}/transactions
[Query]
from: not-a-date
HTTP 400

GET {{host}}/transactions
[Query]
to: not-a-date
HTTP 400
