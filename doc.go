package main

/*

Notes:

x = all seeds that any players have done before
y = total seeds

x / y


1 - (availbe seeds / total seeds of the combined player count)

/*
Notes:

Find lines of code:
find . -type f | grep -v LICENSE | grep -v '/.git/' | grep -v '.png' | grep -v '.jpg' | grep -v gitignore | grep -v certbot-auto | grep -v '.min.css' | grep -v '.min.js' | xargs cat | wc -l

gameState
{
   "turn":1,
   "deckLeft":40,
   "hands":[
      {
         "name":"zamiel",
         "playerNumber":1,
         "cards":[
            {
               "card":{
                  "color":"unknown",
                  "number":0
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"unknown",
                  "number":0
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"unknown",
                  "number":0
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"unknown",
                  "number":0
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"unknown",
                  "number":0
               },
               "clues":[

               ]
            }
         ]
      },
      {
         "name":"zamiel2",
         "playerNumber":2,
         "cards":[
            {
               "card":{
                  "color":"green",
                  "number":4
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"yellow",
                  "number":1
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"red",
                  "number":1
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"red",
                  "number":4
               },
               "clues":[

               ]
            },
            {
               "card":{
                  "color":"yellow",
                  "number":2
               },
               "clues":[

               ]
            }
         ]
      }
   ],
   "playPile":{
      "blue":{
         "card":{
            "color":"",
            "number":0
         },
         "clues":null
      },
      "green":{
         "card":{
            "color":"",
            "number":0
         },
         "clues":null
      },
      "purple":{
         "card":{
            "color":"",
            "number":0
         },
         "clues":null
      },
      "red":{
         "card":{
            "color":"",
            "number":0
         },
         "clues":null
      },
      "yellow":{
         "card":{
            "color":"",
            "number":0
         },
         "clues":null
      }
   },
   "discardPile":[

   ],
   "clues":8,
   "strikes":0,
   "clueHistory":[

   ]
}

*/
