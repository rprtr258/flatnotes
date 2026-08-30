import * as process from "process";

const host = "localhost:8080";

const get_token = await fetch(`${host}/api/token`, {
  method: "POST",
  body: JSON.stringify({
    "username": "user",
    "password": "pass"
  }),
});
const TOKEN = (await get_token.json()).access_token;

switch (process.argv[1]) {
  case "token":
    console.log(TOKEN);
  case "search-all": {
    const result = await fetch(`${host}/api/search?term=*`, {
      headers: {
        Authorization: `Bearer ${TOKEN}`,
      },
    });
    console.log(await result.json());
  }
  case "search-term": {
    const result = await fetch(`${host}/api/search?term=play`, {
      headers: {
        Authorization: `Bearer ${TOKEN}`,
      },
    });
    console.log(await result.json());
  }
  case "tags": {
    const result = await fetch(`${host}/api/tags`, {
      headers: {
        Authorization: `Bearer ${TOKEN}`,
      },
    });
    console.log(await result.json());
  }
  default:
    console.log("unknown command");
    process.exit(1);
}
