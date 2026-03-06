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

  return (
    <>
      <style>{`
        .DocSearch-Sidepanel-Title { display: none; }
        .DocSearch-SidepanelButton.floating {
          background-color: var(--secondary-btn-color);
          width: 56px;
          height: 56px;
          border-radius: 50%;
        }
        .DocSearch-SidepanelButton.floating:hover,
        .DocSearch-SidepanelButton.floating:focus {
          background-color: var(--hover-color) !important;
        }
        .DocSearch-Sidepanel-Prompt--submit,
        .DocSearch-AskAiScreen-RelatedSources-Item-Link {
          width: 24px;
          height: 24px;          
          background-color: var(--border-color) !important;
        }
        .DocSearch-Sidepanel-Content, 
        .DocSearch-AskAiScreen-Response, 
        .DocSearch-Sidepanel-Prompt, 
        .DocSearch-Sidepanel-Header {
          background-color: var(--bg-color);
          scrollbar-color: var(--border-color) var(--bg-color);
        }
        .DocSearch-Sidepanel-Prompt { border-top: 1px solid var(--border-color); }
        .DocSearch-Sidepanel--powered-by {display: none;}
        .DocSearch-Sidepanel-Header { border-bottom: 1px solid var(--border-color); }
        .DocSearch-Sidepanel-Prompt--form {border: 1px solid var(--border-color); }
        .DocSearch-Sidepanel-Prompt--form:hover {border: 1px solid var(--hover-color); }
        .DocSearch-Sidepanel-Prompt--form:active {border: 1px solid var(--hover-color); }
        .DocSearch-Sidepanel-Prompt--form:focus-within {border: 1px solid var(--hover-color); }
        .DocSearch-Sidepanel-SuggestedQuestion { border: 1px solid var(--border-color); }
        .DocSearch-Sidepanel-Footer { background-color: var(--bg-color); }
        .DocSearch-Sidepanel-Prompt--disclaimer { color: var(--secondary-text-color); }
        .DocSearch-Markdown-Content code, .DocSearch-Markdown-Content pre, .DocSearch-CodeSnippet-CopyButton { background-color: var(--block-bg-color); color: var(--primary-text-color); }
        .DocSearch-Sidepanel-RecentConversation { background-color: var(--bg-color); }
        .DocSearch-Sidepanel-Prompt--submit {background-color: var(--border-color) !important;}
        .DocSearch-SidepanelButton.floating svg { width: 24px; height: 24px; }
        .DocSearch-Back-Icon, .DocSearch-Action:hover {color: var(--primary-text-color);}
        .DocSearch-Menu-item, .DocSearch-Menu-item:hover, .DocSearch-Menu-content open {background-color: var(--bg-color);}
        .DocSearch-Markdown-Content li {color: var(--secondary-text-color);}
        .DocSearch-Markdown-Content a {font-size: 0.875rem;}
        .DocSearch-AskAiScreen-RelatedSources-Item-Link, .DocSearch-AskAiScreen-RelatedSources-Item-Link:hover {background-color: var(--block-bg-color); color: var(--secondary-text-color);}
        .DocSearch-Hit-icon-sparkles, .DocSearch-Sidepanel-Header-TitleIcon {color: var(--primary-text-color);}
        .DocSearch-AskAiScreen-RelatedSources {display: none;}
      `}</style>
      <Sidepanel {...(sidepanelProps as any)} />
    </>
  )
}
