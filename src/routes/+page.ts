import { error } from "@sveltejs/kit";
import type { Content } from "@/types";

export function load({ params }) {
  return {
    watched: [] as Content[]
  };
}
