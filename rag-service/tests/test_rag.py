import pytest
from unittest.mock import patch, mock_open, MagicMock
from rag.rag import make_prompt, classifier
from langchain.prompts import ChatPromptTemplate

def test_make_prompt():
    # Define the content of the mock files
    mock_files = {
        'tests/prompts/base.txt': 'base_prompt',
        'tests/prompts/1.txt': 'label1_prompt',
        'tests/prompts/context.txt': 'context_prompt',
        'tests/prompts/lk_context.txt': 'lk_context_prompt',
        'tests/prompts/history_block.txt': 'history_block_prompt',
        'tests/prompts/postfix.txt': 'postfix_prompt',
    }

    # side effect function for mock_open
    def open_side_effect(file, mode='r'):
        for path, content in mock_files.items():
            if file == path:
                return mock_open(read_data=content).return_value
        raise FileNotFoundError(f"File not found: {file}")

    # Patch the open function to return the mock file content
    with patch('builtins.open', side_effect=open_side_effect):
        # Call the function to be tested
        prompt_template = make_prompt(prompts_path='tests/prompts', label='1')

        # Assert that the prompt template is of the correct type
        assert isinstance(prompt_template, ChatPromptTemplate)

        # Assert that the template string is what you expect
        expected_template = "base_promptlabel1_promptcontext_promptlk_context_prompthistory_block_promptpostfix_prompt"
        assert prompt_template.messages[0].prompt.template == expected_template

def test_classifier():
    # Mock the OpenAI client and its response
    mock_response = MagicMock()
    mock_response.choices[0].message.content = "1. Жалоба пользователя"
    
    with patch('rag.rag.OpenAI') as mock_openai:
        mock_openai.return_value.chat.completions.create.return_value = mock_response
        
        # Call the function to be tested
        result = classifier(key="fake_api_key", query="some user query")
        
        # Assert the result
        assert result == "1. Жалоба пользователя"
