import { clsx, } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs) {
	return twMerge(clsx(inputs));
}

export function fmtDate(ts) {
	if (!ts) return null
	const d = new Date(ts.includes('T') ? ts : ts + 'T00:00:00')
	const day = String(d.getDate()).padStart(2, '0')
	const month = String(d.getMonth() + 1).padStart(2, '0')
	return `${day}-${month}-${d.getFullYear()}`
}

export function fmtTs(ts) {
	if (!ts) return null
	const d = new Date(ts)
	const day = String(d.getDate()).padStart(2, '0')
	const month = String(d.getMonth() + 1).padStart(2, '0')
	const hh = String(d.getHours()).padStart(2, '0')
	const mm = String(d.getMinutes()).padStart(2, '0')
	return `${day}-${month}-${d.getFullYear()} ${hh}:${mm}`
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any