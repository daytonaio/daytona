import '@docsearch/css/dist/sidepanel.css'
import '@docsearch/css/dist/style.css'
import { DocSearchSidepanel as Sidepanel } from '@docsearch/react/sidepanel'

const DOCSEARCH_APP_ID = import.meta.env.PUBLIC_DOCSEARCH_APP_ID || null
const DOCSEARCH_API_KEY = import.meta.env.PUBLIC_DOCSEARCH_API_KEY || null
const DOCSEARCH_INDEX_NAME = import.meta.env.PUBLIC_DOCSEARCH_INDEX_NAME || null
const DOCSEARCH_ASSISTANT_ID =
  import.meta.env.PUBLIC_DOCSEARCH_ASSISTANT_ID || null

export default function DocSearchSidepanel() {
  if (!DOCSEARCH_APP_ID || !DOCSEARCH_API_KEY || !DOCSEARCH_INDEX_NAME) {
    return null
  }

  const sidepanelProps = {
    appId: DOCSEARCH_APP_ID,
    apiKey: DOCSEARCH_API_KEY,
    indexName: DOCSEARCH_INDEX_NAME,
    assistantId: DOCSEARCH_ASSISTANT_ID || undefined,
    suggestedQuestions: true,
  }

  return <Sidepanel {...(sidepanelProps as any)} />
}
