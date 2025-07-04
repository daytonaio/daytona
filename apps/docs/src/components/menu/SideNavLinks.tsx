import { useEffect, useState } from 'react'

const GITHUB_STARS_FORMATTER = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1,
})

export const SideNavLinks = () => {
  const [stars, setStars] = useState<number | null | undefined>(undefined)

  useEffect(() => {
    const storedStars = sessionStorage.getItem('stargazers')
    if (!storedStars || isNaN(Number(storedStars))) {
      fetch('https://api.github.com/repos/daytonaio/daytona')
        .then(response => response.json())
        .then(data => {
          setStars(data.stargazers_count)
          sessionStorage.setItem('stargazers', String(data.stargazers_count))
        })
        .catch(error => {
          console.error(error)
          setStars(null)
        })
    } else {
      setStars(Number(storedStars))
    }
  }, [])

  return (
    <>
      <div className="nav-item call">
        <a
          href="https://www.daytona.io/contact"
          target="_blank"
          className="nav__link"
          rel="noreferrer"
        >
          Get a Demo
        </a>
      </div>
      <div className="nav-item github">
        <a
          href="https://github.com/daytonaio"
          target="_blank"
          className="nav__link"
          rel="noreferrer"
        >
          <svg
            width="17"
            height="16"
            viewBox="0 0 17 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              fillRule="evenodd"
              clipRule="evenodd"
              d="M8.86217 0C4.36603 0 0.724365 3.67055 0.724365 8.20235C0.724365 11.8319 3.05381 14.8975 6.28859 15.9843C6.69548 16.0561 6.84807 15.81 6.84807 15.5947C6.84807 15.3999 6.83789 14.754 6.83789 14.067C4.79327 14.4464 4.26431 13.5646 4.10156 13.1033C4.01001 12.8674 3.61329 12.1395 3.26743 11.9447C2.98261 11.7909 2.57572 11.4115 3.25726 11.4013C3.89811 11.391 4.35586 11.9959 4.50845 12.242C5.24085 13.4826 6.41066 13.134 6.87858 12.9187C6.94979 12.3856 7.16341 12.0267 7.39737 11.8216C5.58671 11.6166 3.69467 10.9091 3.69467 7.77173C3.69467 6.87972 4.01001 6.14151 4.52879 5.56735C4.44741 5.36229 4.16259 4.52155 4.61017 3.39372C4.61017 3.39372 5.29171 3.17841 6.84807 4.23446C7.49909 4.04991 8.19081 3.95763 8.88252 3.95763C9.57423 3.95763 10.2659 4.04991 10.917 4.23446C12.4733 3.16816 13.1549 3.39372 13.1549 3.39372C13.6024 4.52155 13.3176 5.36229 13.2362 5.56735C13.755 6.14151 14.0704 6.86947 14.0704 7.77173C14.0704 10.9194 12.1682 11.6166 10.3575 11.8216C10.6525 12.078 10.9068 12.5701 10.9068 13.3391C10.9068 14.4361 10.8966 15.3179 10.8966 15.5947C10.8966 15.81 11.0492 16.0664 11.4561 15.9843C14.6705 14.8975 17 11.8216 17 8.20235C17 3.67055 13.3583 0 8.86217 0Z"
            ></path>
          </svg>
          {stars === undefined
            ? ''
            : stars === null
              ? 'Star'
              : GITHUB_STARS_FORMATTER.format(stars)}
        </a>
      </div>
    </>
  )
}
