import { getPolicies, type Policy } from "@/lib/api";
import Link from "next/link";

// Money inputted as integer cents - conver to dollars only
function formatCentsToDollars(cents: number): string {
    return new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
    }).format(cents / 100);
}

// Set up badge per status, if missing, return error
const statusColors: Record<Policy["status"], string> = {
    quoted: "bg-gray-100 text-gray-800",
    pending: "bg-yellow-100 text-yellow-800",
    active: "bg-green-100 text-green-800",
    past_due: "bg-orange-100 text-orange-800",
    cancelled: "bg-red-100 text-red-800",
    lapsed: "bg-red-100 text-red-800",
    reinstated: "bg-blue-100 text-blue-800",
};

export default async function PoliciesPage() {
    const policies = await getPolicies(); // calls Go directly
    // return a <table> ... policies.map((p) => <tr key={p.id}>...</tr>) ...

    return (
        <main className="mx-auto w-full max-w-4xl p-8">
            <h1 className="mb-6 text-2xl font-semibold">Policies</h1>

            {policies.length === 0 ? ( // if 0, return none
                <p className="text-gray-500">No policies yet.</p>
            ) : ( // if policies, return table
                <table className="w-full text-left text-sm">
                    <thead className="border-b">
                        <tr>
                            <th className="py-2">Unit</th>
                            <th>Status</th>
                            <th className="text-right">Premium</th>
                            <th className="pl-6">Effective</th>
                        </tr>
                    </thead>
                    <tbody>
                        {policies.map((p) => (
                            <tr key={p.id} className="border-b">
                                <td className="py-2">
                                    <Link href={`/policies/${p.id}`} className="text-blue-600 hover:underline">
                                        {p.unit}
                                    </Link>
                                </td>
                                <td>
                                    <span className={`rounded px-2 py-0.5 text-xs font-medium ${statusColors[p.status]}`}>
                                        {p.status}
                                    </span>
                                </td>
                                <td className="text-right">{formatCentsToDollars(p.premium_cents)}</td>
                                <td className="pl-6">{p.effective_date.slice(0, 10)}</td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            )}
        </main>
    ); // for date, keep only yyyy-mm-dd, slice(0, 10) to remove time
}
