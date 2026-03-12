package openai

const partsLinkingPrompt = `
You are a document analysis assistant.
You will be given an ordered list of document chunks extracted from a single document.
Each chunk is numbered and has a role (title, sectionHeading, content, table) and text content.

Analyze all chunks in order and group them by topic.
Each group should contain the indexes of the relevant chunks.

Return a JSON object matching the Document schema.
`

const partsEnrichPrompt = `
You are a document analysis assistant. 
You will be given document chunks extracted from a file, and you must return structured search metadata based on the content.

Each chunk has a role (title, sectionHeading, content, table), page number, section path, and text content.
Analyze all chunks together as a single coherent document.

Return a JSON object with the following fields:
- title: the document title
- summary: a 2-3 sentence summary of the entire document
- keywords: a list of key terms and concepts
- category: the document category or domain
- language: the detected language (e.g. "ko", "en")
`
