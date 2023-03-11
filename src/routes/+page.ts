import { error } from "@sveltejs/kit";
import type { Content } from "@/types";

export function load({ params }) {
  return {
    watched: [
      {
        title: "The Last Of Us",
        poster:
          "https://m.media-amazon.com/images/M/MV5BZGUzYTI3M2EtZmM0Yy00NGUyLWI4ODEtN2Q3ZGJlYzhhZjU3XkEyXkFqcGdeQXVyNTM0OTY1OQ@@._V1_QL75_UX140_CR0,0,140,207_.jpg"
      },
      {
        title: "The Whale",
        poster:
          "https://m.media-amazon.com/images/M/MV5BZDQ4Njg4YTctNGZkYi00NWU1LWI4OTYtNmNjOWMyMjI1NWYzXkEyXkFqcGdeQXVyMTA3MDk2NDg2._V1_QL75_UY207_CR3,0,140,207_.jpg"
      },
      {
        title: "Creed III",
        poster:
          "https://m.media-amazon.com/images/M/MV5BYWY1ZDY4MmQtYjhiYS00N2QwLTk1NzgtOWI2YzUwZThjNDYwXkEyXkFqcGdeQXVyMDM2NDM2MQ@@._V1_QL75_UX140_CR0,0,140,207_.jpg"
      },
      {
        title: "Upright",
        poster:
          "https://m.media-amazon.com/images/M/MV5BNzk5ODJiOWUtMTBkNi00MTU4LWE5YmEtOGMxYmE3YTgwMDZlXkEyXkFqcGdeQXVyMTQyOTE3ODk1._V1_QL75_UY281_CR2,0,190,281_.jpg"
      },
      {
        title: "Interstellar",
        poster:
          "https://m.media-amazon.com/images/M/MV5BZjdkOTU3MDktN2IxOS00OGEyLWFmMjktY2FiMmZkNWIyODZiXkEyXkFqcGdeQXVyMTMxODk2OTU@._V1_QL75_UX190_CR0,0,190,281_.jpg"
      },
      {
        title: "Game of Thrones",
        poster:
          "https://m.media-amazon.com/images/M/MV5BYTRiNDQwYzAtMzVlZS00NTI5LWJjYjUtMzkwNTUzMWMxZTllXkEyXkFqcGdeQXVyNDIzMzcwNjc@._V1_QL75_UY281_CR8,0,190,281_.jpg"
      },
      {
        title: "House of the Dragon",
        poster:
          "https://m.media-amazon.com/images/M/MV5BZjBiOGIyY2YtOTA3OC00YzY1LThkYjktMGRkYTNhNTExY2I2XkEyXkFqcGdeQXVyMTEyMjM2NDc2._V1_QL75_UX190_CR0,0,190,281_.jpg"
      },
      {
        title: "The Punisher",
        poster:
          "https://m.media-amazon.com/images/M/MV5BNjJhZDZhNWYtMjdhYS00NjkyLWE5NzItMzljNmQ3NGE4MGZjXkEyXkFqcGdeQXVyMjkwOTAyMDU@._V1_QL75_UY281_CR11,0,190,281_.jpg"
      },
      {
        title: "Fargo",
        poster:
          "https://m.media-amazon.com/images/M/MV5BN2NiMGE5M2UtNWNlNC00N2Y4LTkwOWUtMDlkMzEwNTcyOTcyXkEyXkFqcGdeQXVyMTkxNjUyNQ@@._V1_QL75_UY281_CR2,0,190,281_.jpg"
      },
      {
        title: "Ash vs Evil Dead",
        poster:
          "https://m.media-amazon.com/images/M/MV5BMTYyMjQyNTE5MF5BMl5BanBnXkFtZTgwMjEyMjE2NDM@._V1_QL75_UX190_CR0,2,190,281_.jpg"
      },
      {
        title: "Bonnie and Clyde",
        poster:
          "https://m.media-amazon.com/images/M/MV5BMTNjNzBlY2QtNmY1Ni00MzhkLThmODgtMzc3ZDQ0YzJjZjNlXkEyXkFqcGdeQXVyMjUzOTY1NTc@._V1_QL75_UX190_CR0,3,190,281_.jpg"
      }
    ] as Content[]
  };
}
