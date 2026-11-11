export function InviteAcceptCardSkeleton() {
	return (
		<div className="flex w-full flex-col items-center text-center" aria-busy="true" aria-hidden>
			<div className="mb-5 size-[5.5rem] animate-pulse rounded-2xl bg-muted sm:mb-6 sm:size-24" />
			<div className="mb-2 h-4 w-40 max-w-[70%] animate-pulse rounded-md bg-muted" />
			<div className="mb-5 h-8 w-48 max-w-[85%] animate-pulse rounded-md bg-muted sm:mb-6" />
			<div className="mb-5 h-5 w-44 animate-pulse rounded-md bg-muted sm:mb-6" />
			<div className="h-10 w-full animate-pulse rounded-[10px] bg-muted sm:h-11 sm:rounded-xl" />
		</div>
	);
}
