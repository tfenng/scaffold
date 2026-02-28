import { NavHeader } from "@/components/layout/nav-header";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen flex flex-col">
      <NavHeader />
      <main className="flex-1 container mx-auto py-6 px-4">
        {children}
      </main>
    </div>
  );
}
