import type { Metadata } from "next";
import { DataModelViewer } from "@/features/dev/data-model/data-model-viewer";
import { loadDataModelDiagrams } from "@/features/dev/data-model/lib/load-data-model-diagrams";

export const metadata: Metadata = {
	title: "Data model — WiChat",
	description: "Simple ERD viewer (Mermaid) for implemented and planned schemas.",
};

export default function DevDataModelPage() {
	const diagrams = loadDataModelDiagrams();
	return (
		<DataModelViewer
			implemented={diagrams.implemented}
			planned={diagrams.planned}
			docsPath={diagrams.docsPath}
		/>
	);
}
