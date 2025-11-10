import React, { useState, useEffect, useRef } from 'react';
import Editor from '@monaco-editor/react';
import { Box, CircularProgress } from '@mui/material';
import { getMonacoLanguage } from '../utils/fileUtils';

const CodeEditor = ({ file, content, readOnly = false, onChange, onSave }) => {
  const [editorContent, setEditorContent] = useState(content || '');
  const [isModified, setIsModified] = useState(false);
  const editorRef = useRef(null);
  const language = getMonacoLanguage(file?.filename);
  const preservedStateRef = useRef(null);

  useEffect(() => {
    if (content !== undefined && editorRef.current) {
      const editor = editorRef.current;
      const currentValue = editor.getValue();
      
      // Only update if content actually changed
      if (currentValue !== content) {
        // Save cursor position, selection, and scroll position BEFORE updating
        const position = editor.getPosition();
        const scrollTop = editor.getScrollTop();
        const scrollLeft = editor.getScrollLeft();
        const selection = editor.getSelection();
        
        // Store state to restore after content update
        preservedStateRef.current = {
          position,
          scrollTop,
          scrollLeft,
          selection,
        };
        
        // Update editor content - this will trigger onChange
        setEditorContent(content);
        setIsModified(false);
      } else if (editorContent !== content) {
        // Content matches but state is out of sync
        setEditorContent(content);
        setIsModified(false);
      }
    } else if (content !== undefined && editorContent !== content) {
      // Editor not mounted yet, just update state
      setEditorContent(content);
      setIsModified(false);
    }
  }, [content, file?.id]);

  useEffect(() => {
    if (editorRef.current && preservedStateRef.current && editorContent) {
      const { position, scrollTop, scrollLeft, selection } = preservedStateRef.current;
      
      const timeoutId = setTimeout(() => {
        if (editorRef.current) {
          try {
            // Restore scroll position first
            editorRef.current.setScrollTop(scrollTop);
            editorRef.current.setScrollLeft(scrollLeft);
            
            // Restore cursor position
            if (position) {
              // Validate position is within document bounds
              const model = editorRef.current.getModel();
              if (model) {
                const lineCount = model.getLineCount();
                const validLine = Math.min(position.lineNumber, lineCount);
                const validColumn = Math.min(
                  position.column,
                  model.getLineMaxColumn(validLine)
                );
                editorRef.current.setPosition({
                  lineNumber: validLine,
                  column: validColumn,
                });
                editorRef.current.revealLineInCenter(validLine);
              }
            }
            
            // Restore selection if it exists
            if (selection) {
              editorRef.current.setSelection(selection);
            }
            
            // Focus editor
            editorRef.current.focus();
          } catch (error) {
            // Silently fail if restoration fails
          }
        }
        // Clear preserved state
        preservedStateRef.current = null;
      }, 10);
      
      return () => clearTimeout(timeoutId);
    }
  }, [editorContent]);

  const handleEditorChange = (value) => {
    setEditorContent(value || '');
    setIsModified(value !== content);
    if (onChange) {
      onChange(value || '');
    }
  };

  const handleEditorDidMount = (editor, monaco) => {
    editorRef.current = editor;
    
    // Add save shortcut (Ctrl+S / Cmd+S)
    if (onSave && !readOnly) {
      editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
        const currentContent = editor.getValue();
        onSave(currentContent);
      });
    }
  };

  return (
    <Box sx={{ width: '100%', height: '100%', minHeight: '400px', position: 'relative' }}>
      <Editor
        height="100%"
        language={language}
        value={editorContent}
        onChange={handleEditorChange}
        onMount={handleEditorDidMount}
        theme="vs-dark"
        options={{
          readOnly: readOnly,
          minimap: { enabled: true },
          fontSize: 14,
          wordWrap: 'on',
          automaticLayout: true,
          scrollBeyondLastLine: false,
          lineNumbers: 'on',
          renderLineHighlight: 'all',
          selectOnLineNumbers: true,
          roundedSelection: false,
          cursorStyle: 'line',
          fontFamily: 'Consolas, "Courier New", monospace',
          tabSize: 2,
          insertSpaces: true,
          detectIndentation: true,
          formatOnPaste: true,
          formatOnType: true,
          suggestOnTriggerCharacters: true,
          acceptSuggestionOnEnter: 'on',
          snippetSuggestions: 'top',
        }}
        loading={
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
            <CircularProgress />
          </Box>
        }
      />
      {isModified && !readOnly && (
        <Box
          sx={{
            position: 'absolute',
            top: 8,
            left: 8,
            backgroundColor: 'rgba(255, 193, 7, 0.95)',
            color: 'black',
            px: 2,
            py: 0.75,
            borderRadius: 1,
            fontSize: '0.75rem',
            fontWeight: 600,
            boxShadow: '0 2px 8px rgba(0, 0, 0, 0.2)',
            zIndex: 10,
            display: 'flex',
            alignItems: 'center',
            gap: 1,
            maxWidth: 'calc(100% - 16px)',
            border: '1px solid rgba(255, 193, 7, 0.3)',
          }}
        >
          <Box
            sx={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              backgroundColor: '#ff9800',
              animation: 'pulse 2s infinite',
              '@keyframes pulse': {
                '0%, 100%': {
                  opacity: 1,
                },
                '50%': {
                  opacity: 0.5,
                },
              },
            }}
          />
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Box sx={{ fontWeight: 600, mb: 0.25 }}>Kaydedilmemiş değişiklikler</Box>
            <Box sx={{ fontSize: '0.7rem', opacity: 0.8 }}>Ctrl+S ile kaydet</Box>
          </Box>
        </Box>
      )}
    </Box>
  );
};

export default CodeEditor;

