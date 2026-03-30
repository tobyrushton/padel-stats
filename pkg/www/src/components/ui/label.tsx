import * as React from "react"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"

function Label({
  className,
  asChild = false,
  ...props
}: React.ComponentProps<"label"> & {
  asChild?: boolean
}) {
  const Comp = asChild ? Slot.Root : "label"

  return (
    <Comp
      data-slot="label"
      className={cn(
        "flex items-center gap-2 text-sm leading-none font-medium select-none",
        className,
      )}
      {...props}
    />
  )
}

export { Label }
