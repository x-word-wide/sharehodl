import { Button, Card, CardContent, CardHeader, CardTitle } from "@repo/ui";
import { BlockList } from "./components/BlockList";
import { TransactionList } from "./components/TransactionList";
import { Search, Activity, Box, Users, Zap } from "lucide-react";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        {/* Hero Section */}
        <section className="mb-12 text-center">
          <h1 className="mb-4 text-4xl font-extrabold tracking-tight lg:text-5xl">
            ShareHODL Blockchain Explorer
          </h1>
          <p className="mb-8 text-lg text-muted-foreground">
            Search blocks, transactions, and addresses on the ShareHODL network.
          </p>
          <div className="mx-auto max-w-2xl relative">
            <Search className="absolute left-3 top-3 h-5 w-5 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search by Hodl Address (HodlXX...) / Tx Hash / Block / Token"
              className="w-full rounded-lg border bg-background py-3 pl-10 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
        </section>

        {/* Network Stats */}
        <section className="mb-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Market Cap
              </CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">$12.5M</div>
              <p className="text-xs text-muted-foreground">
                +2.5% from last month
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Block Height
              </CardTitle>
              <Box className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">1,234,567</div>
              <p className="text-xs text-muted-foreground">
                +1 every 2s
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Active Validators
              </CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">120</div>
              <p className="text-xs text-muted-foreground">
                +4 new this week
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Avg Block Time
              </CardTitle>
              <Zap className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">2.1s</div>
              <p className="text-xs text-muted-foreground">
                Network is healthy
              </p>
            </CardContent>
          </Card>
        </section>

        {/* Lists */}
        <div className="grid gap-8 md:grid-cols-2">
          <BlockList />
          <TransactionList />
        </div>
      </main>
    </div>
  );
}
