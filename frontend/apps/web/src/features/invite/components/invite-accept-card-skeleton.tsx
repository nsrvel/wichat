export function InviteAcceptCardSkeleton() {
	return (
		<div className="flex w-full flex-col items-center text-center" aria-busy="true" aria-hidden>
			<div className="mb-4 size-[4.5rem] animate-pulse rounded-2xl bg-muted sm:mb-5 sm:size-20" />
			<div className="mb-1.5 h-3.5 w-36 max-w-[65%] animate-pulse rounded-md bg-muted" />
			<div className="mb-3 h-7 w-44 max-w-[80%] animate-pulse rounded-md bg-muted" />
			<div className="mb-5 h-6 w-40 animate-pulse rounded-full bg-muted" />
			<div className="h-10 w-full animate-pulse rounded-[10px] bg-muted sm:h-11 sm:rounded-xl" />
		</div>
	);
}
