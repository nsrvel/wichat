import fs from "node:fs";
import path from "node:path";

const DATA_MODEL_DIR = path.join(process.cwd(), "../../../docs/architecture/data-model");

function readDiagram(filename: string): string | null {
	const filePath = path.join(DATA_MODEL_DIR, filename);
	try {
		return fs.readFileSync(filePath, "utf8");
	} catch {
		return null;
	}
}

export type DataModelDiagramSet = {
	implemented: string | null;
	planned: string | null;
	docsPath: string;
};

/** Reads Mermaid sources from repo `docs/architecture/data-model/`. */
export function loadDataModelDiagrams(): DataModelDiagramSet {
	return {
		implemented: readDiagram("implemented-auth.mermaid"),
		planned: readDiagram("planned-overview.mermaid"),
		docsPath: "docs/architecture/data-model/",
	};
}
