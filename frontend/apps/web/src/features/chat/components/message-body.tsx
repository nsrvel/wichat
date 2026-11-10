import { cn } from "@/lib/utils";

type MessageBodyProps = {
	text: string;
	className?: string;
};

/** Plain-text body — one or more paragraphs split on blank lines. */
export function MessageBody({ text, className }: MessageBodyProps) {
	const paragraphs = text.split(/\n\n+/).filter((p) => p.trim().length > 0);
	if (paragraphs.length === 0) {
		return null;
	}
	return (
		<>
			{paragraphs.map((paragraph, index) => (
				<p key={index} className={cn(index > 0 && "mt-[9px]", className)}>
					{paragraph}
				</p>
			))}
		</>
	);
}
