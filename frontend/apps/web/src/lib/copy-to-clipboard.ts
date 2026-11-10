/** Copies text; prefers Clipboard API with a textarea fallback for strict gesture policies. */
export async function copyTextToClipboard(text: string): Promise<boolean> {
	const value = text.trim();
	if (!value) {
		return false;
	}

	if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
		try {
			await navigator.clipboard.writeText(value);
			return true;
		} catch {
			// fall through
		}
	}

	if (typeof document === "undefined") {
		return false;
	}

	const textarea = document.createElement("textarea");
	textarea.value = value;
	textarea.setAttribute("readonly", "");
	textarea.style.position = "fixed";
	textarea.style.left = "-9999px";
	document.body.appendChild(textarea);
	textarea.select();
	textarea.setSelectionRange(0, value.length);

	let ok = false;
	try {
		ok = document.execCommand("copy");
	} catch {
		ok = false;
	}

	document.body.removeChild(textarea);

	return ok;
}
