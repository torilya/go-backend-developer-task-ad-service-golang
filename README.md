# Тестовое задание

Нужно написать сервис на Go, который принимает на вход HTTP GET запрос вида `/ad?client=1&slot=1&user=1`

Сервис формирует bidRequest.json из шаблона и сохраняет его в файл
Берет из JSON-шаблона bid-response.json строку `adm` и отдает ее в ответ на запрос

bid-request.json
```json
{
  "id": "{{ RANDOM_UUID }}",
  "site": {
    "id": "{{ CLIENT_ID }}",
    // значение из поля client во входящем HTTP GET запросе
    "ref": "http://example.com",
    "publisher": {
      "name": "example.com",
      "id": "{{ CLIENT_ID }}"
      // значение из поля client во входящем HTTP GET запросе
    },
    "name": "{{ CLIENT_ID }}"
    // значение из поля client во входящем HTTP GET запросе
  },
  "wseat": [
    "{{ SLOT_ID }}"
    // значение из поля slot во входящем HTTP GET запросе
  ],
  "user": {
    "id": "{{ USER_ID }}"
    // значение из поля user во входящем HTTP GET запросе
  },
  "device": {
    "language": "ru",
    "geo": {
      "country": "RU"
    },
    "ip": "{{ IP_FROM_INCOMMING_REQUEST }}"
    // IP откуда получили HTTP GET запрос
  },
  "tmax": 75,
  "cur": [
    "USD"
  ],
  "imp": [
    {
      "bidfloor": 3.213,
      "id": "1",
      "banner": {
        "pos": 1,
        "h": 600,
        "w": 600,
        "format": [
          {
            "h": 300,
            "w": 300
          }
        ]
      }
    }
  ],
  "at": 1
}
```

bid-response.json
```json
{
  "id": "IxexyLDIIk",
  "seatbid": [
    {
      "bid": [
        {
          "id": "1",
          "impid": "1",
          "price": 0.751371,
          "burl": "http://ads.example.com/win/12345",
          "adm": "<a href=\"http://ya.ru/\"><img src=\"http://via.placeholder.com/600x600\" width=\"600\" height=\"600\" border=\"0\" alt=\"Advertisement\" /></a>",
          "adomain": [
            "example.com"
          ]
        }
      ],
      "seat": "1"
    }
  ],
  "cur": "USD"
}
```

%% EOF