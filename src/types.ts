export interface Content {
  /**
   * Title of content.
   */
  title: string;

  /**
   * URL to poster image.
   */
  poster: string;
}

export interface Watched {
  watched: boolean;
  rating: number;
  content: Content;
}
