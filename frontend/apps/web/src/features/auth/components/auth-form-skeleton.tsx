export function AuthFormSkeleton() {
	return (
		<div className="flex flex-col gap-2.5 sm:gap-3.5" aria-hidden>
			<div className="h-10 rounded-[10px] bg-[#f0f0f0] sm:h-11 sm:rounded-xl dark:bg-white/[0.08]" />
			<div className="h-3 w-8 bg-[#e8e8e8] dark:bg-white/[0.06]" />
			<div className="h-10 rounded-[10px] bg-[#f0f0f0] sm:h-11 sm:rounded-xl dark:bg-white/[0.08]" />
			<div className="h-10 rounded-[10px] bg-[#f0f0f0] sm:h-11 sm:rounded-xl dark:bg-white/[0.08]" />
			<div className="h-10 rounded-[10px] bg-[#f0f0f0] sm:h-11 sm:rounded-xl dark:bg-white/[0.08]" />
		</div>
	);
}
