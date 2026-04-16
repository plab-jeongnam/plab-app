import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"

interface AppState {
  theme: "light" | "dark" | "system"
  sidebarOpen: boolean
}

interface AppActions {
  setTheme: (theme: AppState["theme"]) => void
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
}

type AppStore = AppState & AppActions

export const useAppStore = create<AppStore>()(
  devtools(
    persist(
      (set) => ({
        theme: "system",
        sidebarOpen: true,
        setTheme: (theme) => set({ theme }),
        toggleSidebar: () =>
          set((state) => ({ sidebarOpen: !state.sidebarOpen })),
        setSidebarOpen: (open) => set({ sidebarOpen: open }),
      }),
      {
        name: "app-store",
        partialize: (state) => ({ theme: state.theme }),
      }
    )
  )
)
