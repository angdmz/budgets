import { useEffect, useRef, useCallback, type ReactNode } from 'react';

interface DialogProps {
  titleId?: string;
  title?: string;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
  dismissible?: boolean;
  'data-testid'?: string;
}

export default function Dialog({
  titleId,
  title,
  onClose,
  children,
  footer,
  dismissible = true,
  'data-testid': dataTestId,
}: DialogProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const previouslyFocused = useRef<HTMLElement | null>(null);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape' && dismissible) {
        e.stopPropagation();
        onClose();
        return;
      }
      if (e.key === 'Tab' && panelRef.current) {
        const focusable = panelRef.current.querySelectorAll<HTMLElement>(
          'a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])'
        );
        if (focusable.length === 0) return;
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    },
    [dismissible, onClose]
  );

  useEffect(() => {
    previouslyFocused.current = document.activeElement as HTMLElement;
    document.body.style.overflow = 'hidden';
    document.addEventListener('keydown', handleKeyDown);

    // Move focus into the panel
    if (panelRef.current) {
      const focusable = panelRef.current.querySelector<HTMLElement>(
        'a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])'
      );
      focusable?.focus();
    }

    return () => {
      document.body.style.overflow = '';
      document.removeEventListener('keydown', handleKeyDown);
      previouslyFocused.current?.focus();
    };
  }, [handleKeyDown]);

  const generatedTitleId = titleId || (title ? `dialog-title-${Math.random().toString(36).slice(2, 9)}` : undefined);

  return (
    <div
      className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-end md:items-center justify-center p-0 md:p-4 z-50"
      data-testid={dataTestId}
      onClick={(e) => {
        if (e.target === e.currentTarget && dismissible) onClose();
      }}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={generatedTitleId}
        className="bg-white dark:bg-gray-800 rounded-t-2xl md:rounded-lg w-full md:max-w-md max-h-[90dvh] overflow-y-auto flex flex-col shadow-xl"
      >
        {title && (
          <div className="flex items-center justify-between p-6 pb-0">
            <h2 id={generatedTitleId} className="text-lg font-semibold dark:text-white">
              {title}
            </h2>
            {dismissible && (
              <button
                type="button"
                onClick={onClose}
                aria-label={typeof title === 'string' ? `Close ${title}` : 'Close dialog'}
                className="ml-4 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                  <path d="M18 6L6 18M6 6l12 12" />
                </svg>
              </button>
            )}
          </div>
        )}
        <div className={title ? 'px-6' : 'p-6'}>{children}</div>
        {footer && (
          <div className="sticky bottom-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 p-4 pb-[max(1rem,env(safe-area-inset-bottom))] flex flex-col-reverse gap-2 md:flex-row md:justify-end md:space-x-3">
            {footer}
          </div>
        )}
      </div>
    </div>
  );
}
