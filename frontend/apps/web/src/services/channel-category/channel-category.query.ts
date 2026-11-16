import { queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { Q_CATEGORY_PERMISSIONS } from "@/services/channel-category/channel-category.key";
import { listCategoryPermissions } from "@/services/channel-category/channel-category.service";

export function categoryPermissionsQueryOpts(groupId: string, categoryId: string) {
	return queryOptions({
		queryKey: [Q_CATEGORY_PERMISSIONS, groupId, categoryId],
		queryFn: () => listCategoryPermissions(groupId, categoryId),
		enabled: isPublicEntityId(groupId) && isPublicEntityId(categoryId),
		staleTime: 60_000,
		refetchOnWindowFocus: false,
	});
}
