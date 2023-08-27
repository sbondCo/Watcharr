import { error } from "@sveltejs/kit";

export async function load({ params }) {
  const { id, username } = params;

  if (!id || !username) {
    error(400);
    return;
  }

  return {
    id,
    username
  };
}
