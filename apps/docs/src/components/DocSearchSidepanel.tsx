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
    expandedWidth: '50vw',
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
        @media screen and (max-width: 768px) {
          .DocSearch-Sidepanel-Action-expand { display: none !important; }
          .DocSearch-Action[title="Conversation history"] { display: none !important; }
        }
        .DocSearch-Back-Icon, .DocSearch-Action:hover {color: var(--primary-text-color);}
        .DocSearch-Menu-item, .DocSearch-Menu-item:hover, .DocSearch-Menu-content open {background-color: var(--bg-color);}
        .DocSearch-Markdown-Content li {color: var(--secondary-text-color); font-size: 1rem;}
        .DocSearch-Markdown-Content a {font-size: 1rem;}
        .DocSearch-AskAiScreen-RelatedSources-Item-Link, .DocSearch-AskAiScreen-RelatedSources-Item-Link:hover {background-color: var(--block-bg-color); color: var(--secondary-text-color);}
        .DocSearch-Hit-icon-sparkles, .DocSearch-Sidepanel-Header-TitleIcon {color: var(--primary-text-color);}
        .DocSearch-AskAiScreen-RelatedSources {display: none;}
        .DocSearch-Sidepanel-Screen--introduction {color: var(--secondary-text-color);}
        .DocSearch-Sidepanel-Screen--title {color: var(--primary-text-color);}
        .DocSearch-AskAiScreen-MessageContent-Tool, .Tool--Result, .DocSearch-AskAiScreen-MessageContent-Tool-Query {color: var(--secondary-text-color);}
        .DocSearch-Sidepanel-SuggestedQuestion {color: var(--secondary-text-color);}
        .DocSearch-Sidepanel-SuggestedQuestion:hover {color: var(--primary-text-color); border: 1px solid var(--primary-text-color);}
        .DocSearch-Markdown-Content > h3, .DocSearch-Markdown-Content > h2 {color: var(--primary-text-color);}
        .DocSearch-Sidepanel-Action-expand svg {
          display: none;
        }
        .DocSearch-Sidepanel-Action-expand::before {
          content: '';
          width: 14px;
          height: 14px;
          display: block;
          background-color: currentColor;
          -webkit-mask: var(--docsearch-sidepanel-expand-icon) center / contain no-repeat;
          mask: var(--docsearch-sidepanel-expand-icon) center / contain no-repeat;
        }
        .DocSearch-Sidepanel-Container {
          --docsearch-sidepanel-expand-icon: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 13 13' fill='none'%3E%3Cpath fill-rule='evenodd' clip-rule='evenodd' d='M8 0.5C8 0.223858 8.22386 0 8.5 0H12.5C12.7761 0 13 0.223858 13 0.5V4.5C13 4.77614 12.7761 5 12.5 5C12.2239 5 12 4.77614 12 4.5V1.70711L8.18689 5.52022C7.99162 5.71548 7.67504 5.71548 7.47978 5.52022C7.28452 5.32496 7.28452 5.00838 7.47978 4.81311L11.2929 1H8.5C8.22386 1 8 0.776142 8 0.5ZM5.52022 7.47978C5.71548 7.67504 5.71548 7.99162 5.52022 8.18689L1.70711 12H4.5C4.77614 12 5 12.2239 5 12.5C5 12.7761 4.77614 13 4.5 13H0.5C0.223858 13 0 12.7761 0 12.5V8.5C0 8.22386 0.223858 8 0.5 8C0.776142 8 1 8.22386 1 8.5V11.2929L4.81311 7.47978C5.00838 7.28452 5.32496 7.28452 5.52022 7.47978Z' fill='black'/%3E%3C/svg%3E");
        }
        .DocSearch-Sidepanel-Container.is-expanded {
          --docsearch-sidepanel-expand-icon: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 13 13' fill='none'%3E%3Cpath fill-rule='evenodd' clip-rule='evenodd' d='M8 0.5C8 0.223858 8.22386 0 8.5 0H12.5C12.7761 0 13 0.223858 13 0.5V4.5C13 4.77614 12.7761 5 12.5 5C12.2239 5 12 4.77614 12 4.5V1.70711L8.18689 5.52022C7.99162 5.71548 7.67504 5.71548 7.47978 5.52022C7.28452 5.32496 7.28452 5.00838 7.47978 4.81311L11.2929 1H8.5C8.22386 1 8 0.776142 8 0.5Z' fill='black' transform='translate(-8 8)'/%3E%3Cpath fill-rule='evenodd' clip-rule='evenodd' d='M5.52022 7.47978C5.71548 7.67504 5.71548 7.99162 5.52022 8.18689L1.70711 12H4.5C4.77614 12 5 12.2239 5 12.5C5 12.7761 4.77614 13 4.5 13H0.5C0.223858 13 0 12.7761 0 12.5V8.5C0 8.22386 0.223858 8 0.5 8C0.776142 8 1 8.22386 1 8.5V11.2929L4.81311 7.47978C5.00838 7.28452 5.32496 7.28452 5.52022 7.47978Z' fill='black' transform='translate(8 -8)'/%3E%3C/svg%3E");
        }
      `}</style>
      <Sidepanel {...(sidepanelProps as any)} />
    </>
  )
}
