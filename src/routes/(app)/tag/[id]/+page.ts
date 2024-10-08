import { error } from "@sveltejs/kit";

export const load = async ({ params }) => {
  const { id } = params;

  if (!id) {
    error(400);
    return;
  }

  return {
    tagId: Number(id)
  };
};
