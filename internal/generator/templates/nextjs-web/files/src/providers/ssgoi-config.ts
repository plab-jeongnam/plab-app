import { fade, slide } from "@ssgoi/react/view-transitions"

export const ssgoiConfig = {
  transitions: [
    {
      from: "/",
      to: "/form",
      transition: slide({ direction: "left" }),
    },
    {
      from: "/form",
      to: "/",
      transition: slide({ direction: "right" }),
    },
  ],
  defaultTransition: fade(),
}
