export function getCodingLanguage(language) {
    switch (language) {
        case 'go':
            return 'Go';
        case 'ts':
            return 'TypeScript';
        case 'tsx':
            return 'TypeScript';
        case 'js':
            return 'JavaScript';
        case 'css':
            return 'CSS';
        case 'scss':
            return 'SCSS';
        case 'sass':
            return 'SCSS';
        case 'json':
            return 'JSON';
        case 'html':
            return 'HTML';
        case 'xml':
            return 'XML';
        case 'php':
            return 'PHP';
        case 'cs':
            return 'C#';
        case 'cpp':
            return 'C++';
        case 'cc':
            return 'C++';
        case 'md':
            return 'Markdown';
        case 'markdown':
            return 'Markdown';
        case 'java':
            return 'Java';
        case 'vb':
            return 'VB';
        case 'coffee':
            return 'CoffeeScript';
        case 'py':
            return 'Python';
        case 'rb':
            return 'Ruby';
        case 'r':
            return 'R';
        case 'm':
            return 'Objective-C';
        default:
            return 'Powershell';
    }
}
