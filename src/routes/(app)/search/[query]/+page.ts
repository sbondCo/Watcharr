import { error } from "@sveltejs/kit";

export async function load({ params }) {
  const { query } = params;

  if (!query) {
    error(400);
    return;
  }

  return {
    slug: query
  };
}
