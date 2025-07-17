// Main menu data
const mainmenue = [
    {
        title: "My LinkedIn",
        link: "https://www.linkedin.com/in/daniel-seifert-23164b6a/",
        icon: "linkedin",
        classes: "blue",
        external: true
    },
    {
        title: "My Github",
        link: "https://github.com/dseif0x",
        icon: "github",
        classes: "lime",
        external: true
    }
];

// Create flowlist function - JavaScript equivalent of PHP create_flowlist
function create_flowlist(icons, small = true, center = false) {
    let icst = `<div class="flow-list ${center ? "center-text" : ""}">`;
    
    icons.forEach(ico => {
        const targetBlank = ico.external !== false ? 'target="_blank"' : '';
        const smallClass = small ? "smallt" : "";
        
        icst += `
            <div class="flow-list-item">
                <a title="${ico.title}" data-toggle="tooltip" class="hlink ${smallClass}" ${targetBlank} href="${ico.link}">
                    <i class="${ico.classes} mdi mdi-${ico.icon}"></i>
                </a>
            </div>`;
    });
    
    icst += '</div>';
    return icst;
}

// Initialize the page when DOM is ready
document.addEventListener('DOMContentLoaded', function() {
    // Set current year
    document.getElementById('current-year').textContent = new Date().getFullYear();
    
    // Generate and insert main menu
    const mainMenuContainer = document.getElementById('main-menu');
    if (mainMenuContainer) {
        mainMenuContainer.innerHTML = create_flowlist(mainmenue, false, true);
    }
    
    // Re-initialize Bootstrap tooltips for dynamically added content
    if (typeof $ !== 'undefined') {
        $('[data-toggle="tooltip"]').tooltip();
    }
});