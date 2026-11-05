export function AuthFormSkeleton() {
	return (
		<div className="flex flex-col gap-2.5 sm:gap-3.5" aria-hidden>
			<div className="h-10 rounded-[10px] bg-field sm:h-11 sm:rounded-xl" />
			<div className="h-3 w-8 bg-muted" />
			<div className="h-10 rounded-[10px] bg-field sm:h-11 sm:rounded-xl" />
			<div className="h-10 rounded-[10px] bg-field sm:h-11 sm:rounded-xl" />
			<div className="h-10 rounded-[10px] bg-field sm:h-11 sm:rounded-xl" />
		</div>
	);
}
