import { Navigation } from "@repo/ui";
import { AddressDisplay } from "./components/AddressDisplay";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            <span className="text-2xl">₿</span>
            ShareWallet
          </h1>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Securely manage your ShareHODL assets with enterprise-grade custody.
          </p>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <AddressDisplay />
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-3 mt-8">
          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">
                Portfolio Value
              </h3>
            </div>
            <div>
              <div className="text-2xl font-bold">$25,430</div>
              <p className="text-sm text-muted-foreground">Total balance</p>
            </div>
          </div>

          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">
                Send
              </h3>
            </div>
            <div>
              <button className="w-full bg-blue-500 text-white px-4 py-2 rounded">Transfer Assets</button>
            </div>
          </div>

          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">
                Receive
              </h3>
            </div>
            <div>
              <button className="w-full bg-green-500 text-white px-4 py-2 rounded">Get Address</button>
            </div>
          </div>
        </div>

        <div className="mt-12 p-8 border rounded-lg bg-muted/50 text-center">
          <p className="text-muted-foreground">Wallet Interface Loading...</p>
        </div>
      </main>
    </div>
  );
}