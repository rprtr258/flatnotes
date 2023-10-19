local host = "localhost:8080";

local get_token = |||
    let REMOTE = "%(host)s"
    ---
    POST
    ${REMOTE}/api/token

    {
        "username": "user",
        "password": "pass"
    }
    ---
    let TOKEN = result.access_token
||| % {host: host};

local meta = {
    host: host,
    get_token: get_token,
};

{
    "token.l2": |||
        %(get_token)s
    ||| % meta,
    "search-all.l2": |||
        %(get_token)s
        ---
        GET
        ${REMOTE}/api/search?term=*

        Authorization: 'Bearer ${TOKEN}'
    ||| % meta,
    "search-term.l2": |||
        %(get_token)s
        ---
        GET
        ${REMOTE}/api/search?term=play

        Authorization: 'Bearer ${TOKEN}'
    ||| % meta,
    "tags.l2": |||
        %(get_token)s
        ---
        GET
        ${REMOTE}/api/tags

        Authorization: 'Bearer ${TOKEN}'
    ||| % meta,
}