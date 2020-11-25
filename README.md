![Docker Image CI](https://github.com/alyrot/uksh-menu-parser/workflows/Docker%20Image%20CI/badge.svg)
![Go](https://github.com/alyrot/uksh-menu-parser/workflows/Go/badge.svg)

# uksh-menu-parser
The bistro of the UKSH Lübeck only publishes their lunch menu as a PDF on their [Website](https://www.uksh.de/servicesternnord/Unser+Speisenangebot/Speisepl%C3%A4ne+L%C3%BCbeck/UKSH_Bistro+L%C3%BCbeck-p-346.html).
This project automatically downloads the lunch plan pdfs and offers
their content as json via an REST-API. The price of the meals sometimes glitches a bit as it is extracted via OCR.

## Endpoints
- /alive : Just returns some dummy text and Status Code 200/OK. Can be used to monitor the availability of the service
- /menu/yyy-mm-dd : Returns the menu for the given date as an json array
```
[
  { Title : string
    Description : string
    Price : string
    Kcal : string
    Type : string
    Date : string // ISO 8601
  }
]
```
If the date is malformed, to far in the past (varies depending on the menu pdf availability)
or more than 7 days in the future, 400/BadRequest is returned. If there is any other error 500/InternalServerError is returned.
