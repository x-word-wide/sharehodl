import { Card, CardContent, CardHeader, CardTitle, Button, Header } from "@repo/ui";
import { Box, Users, Search, ShieldCheck } from "lucide-react";
import Link from "next/link";

const MOCK_VALIDATORS = [
    { rank: 1, name: "Validator One", address: "sharevaloper1...", votingPower: "15.2%", uptime: "100%", commission: "5%" },
    { rank: 2, name: "Validator Two", address: "sharevaloper2...", votingPower: "10.5%", uptime: "99.9%", commission: "10%" },
    { rank: 3, name: "Validator Three", address: "sharevaloper3...", votingPower: "8.1%", uptime: "100%", commission: "5%" },
    { rank: 4, name: "Validator Four", address: "sharevaloper4...", votingPower: "7.4%", uptime: "99.8%", commission: "7%" },
    { rank: 5, name: "Validator Five", address: "sharevaloper5...", votingPower: "6.2%", uptime: "100%", commission: "5%" },
    { rank: 6, name: "Validator Six", address: "sharevaloper6...", votingPower: "5.8%", uptime: "99.5%", commission: "1%" },
    { rank: 7, name: "Validator Seven", address: "sharevaloper7...", votingPower: "4.3%", uptime: "100%", commission: "5%" },
    { rank: 8, name: "Validator Eight", address: "sharevaloper8...", votingPower: "3.9%", uptime: "98.2%", commission: "10%" },
    { rank: 9, name: "Validator Nine", address: "sharevaloper9...", votingPower: "2.5%", uptime: "100%", commission: "5%" },
    { rank: 10, name: "Validator Ten", address: "sharevaloper10...", votingPower: "1.2%", uptime: "99.9%", commission: "2%" },
];

export default function Validators() {
    return (
        <div className="min-h-screen bg-background">
            <Header appName="Explorer" />

            <main className="container mx-auto px-4 py-8">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
                    <div>
                        <h1 className="text-3xl font-bold flex items-center gap-3 mb-2">
                            <Users className="h-8 w-8 text-muted-foreground" />
                            Active Validators
                        </h1>
                        <p className="text-muted-foreground">
                            Top 100 validators securing the ShareHODL network.
                        </p>
                    </div>
                    <div className="relative w-full md:w-96">
                        <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                        <input
                            type="text"
                            placeholder="Search validator..."
                            className="w-full rounded-md border bg-background py-2 pl-9 pr-4 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        />
                    </div>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle>Validator Set</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="overflow-x-auto">
                            <table className="w-full text-sm text-left">
                                <thead className="text-muted-foreground border-b">
                                    <tr>
                                        <th className="h-12 px-4 font-medium">Rank</th>
                                        <th className="h-12 px-4 font-medium">Validator</th>
                                        <th className="h-12 px-4 font-medium">Voting Power</th>
                                        <th className="h-12 px-4 font-medium">Uptime</th>
                                        <th className="h-12 px-4 font-medium text-right">Commission</th>
                                        <th className="h-12 px-4 font-medium text-right">Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {MOCK_VALIDATORS.map((val) => (
                                        <tr key={val.rank} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                                            <td className="p-4 font-medium">{val.rank}</td>
                                            <td className="p-4">
                                                <div className="flex items-center gap-3">
                                                    <div className="h-8 w-8 rounded-full bg-secondary flex items-center justify-center text-xs font-bold">
                                                        {val.name.substring(0, 2).toUpperCase()}
                                                    </div>
                                                    <div>
                                                        <div className="font-medium text-primary">{val.name}</div>
                                                        <div className="text-xs text-muted-foreground font-mono">{val.address}</div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="p-4">{val.votingPower}</td>
                                            <td className="p-4">
                                                <div className="flex items-center gap-2">
                                                    <ShieldCheck className="h-4 w-4 text-green-500" />
                                                    {val.uptime}
                                                </div>
                                            </td>
                                            <td className="p-4 text-right">{val.commission}</td>
                                            <td className="p-4 text-right">
                                                <Button variant="outline" size="sm">Delegate</Button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </CardContent>
                </Card>
            </main>
        </div>
    );
}
